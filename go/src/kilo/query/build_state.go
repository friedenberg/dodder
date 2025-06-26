package query

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/box"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/store_workspace"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/tag_blobs"
)

type stackEl interface {
	sku.Query
	Add(sku.Query) error
}

type buildState struct {
	options

	builder      *Builder
	group        *Query
	latentErrors errors.Multi
	missingBlobs []ErrBlobMissing

	luaVMPoolBuilder        *lua.VMPoolBuilder
	pinnedObjectIds         []pinnedObjectId
	pinnedExternalObjectIds []sku.ExternalObjectId
	workspaceStore          store_workspace.Store

	workspaceStoreAcceptedQueryComponent bool

	scanner box.Scanner
}

func (src *buildState) copy() (dst *buildState) {
	dst = &buildState{
		options:      src.options,
		builder:      src.builder,
		latentErrors: errors.MakeMulti(),
	}

	if src.luaVMPoolBuilder != nil {
		dst.luaVMPoolBuilder = src.luaVMPoolBuilder.Clone()
	}

	dst.group = dst.makeGroup()

	dst.pinnedObjectIds = make([]pinnedObjectId, len(src.pinnedObjectIds))
	copy(dst.pinnedObjectIds, src.pinnedObjectIds)

	dst.pinnedExternalObjectIds = make(
		[]sku.ExternalObjectId,
		len(src.pinnedExternalObjectIds),
	)

	copy(dst.pinnedExternalObjectIds, src.pinnedExternalObjectIds)

	return
}

func (b *buildState) makeGroup() *Query {
	return &Query{
		hidden:           b.builder.hidden,
		optimizedQueries: make(map[genres.Genre]*expSigilAndGenre),
		userQueries:      make(map[ids.Genre]*expSigilAndGenre),
		types:            ids.MakeMutableTypeSet(),
	}
}

func (b *buildState) build(
	values ...string,
) (err error, latent errors.Multi) {
	em := errors.MakeMulti()
	latent = em

	var remaining []string

	if b.workspaceStore == nil {
		remaining = values
	} else {
		for _, value := range values {
			if value == "." {
				b.group.dotOperatorActive = true
				remaining = append(remaining, value)
			}

			var externalObjectIds []sku.ExternalObjectId

			if externalObjectIds, err = b.workspaceStore.GetObjectIdsForString(
				value,
			); err != nil {
				if value != "." {
					remaining = append(remaining, value)
				}

				em.Add(err)
				err = nil

				continue
			}

			b.workspaceStoreAcceptedQueryComponent = true

			for _, externalObjectId := range externalObjectIds {
				if externalObjectId.GetGenre() == genres.None {
					err = errors.ErrorWithStackf("id with empty genre: %q", externalObjectId)
					return
				}

				b.pinnedExternalObjectIds = append(
					b.pinnedExternalObjectIds,
					externalObjectId,
				)
			}
		}
	}

	remainingWithSpaces := make([]string, 0, len(remaining)*2)

	for i, s := range remaining {
		if i > 0 {
			remainingWithSpaces = append(remainingWithSpaces, " ")
		}

		remainingWithSpaces = append(remainingWithSpaces, s)
	}

	reader := catgut.MakeMultiRuneReader(remainingWithSpaces...)
	b.scanner.Reset(reader)

	for b.scanner.CanScan() {
		if err = b.parseTokens(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	for _, k := range b.pinnedExternalObjectIds {
		if k.GetGenre() == genres.None {
			err = errors.ErrorWithStackf("id with empty genre: %q", k)
			return
		}

		if err = b.group.addExactExternalObjectId(b, k); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	for _, k := range b.pinnedObjectIds {
		q := b.makeQuery()

		if err = q.addPinnedObjectId(b, k); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = b.group.add(q); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	b.addDefaultsIfNecessary()

	if err = b.group.reduce(b); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (b *buildState) addDefaultsIfNecessary() {
	if b.defaultGenres.IsEmpty() || !b.group.isEmpty() {
		return
	}

	if b.builder.requireNonEmptyQuery && b.group.isEmpty() {
		return
	}

	if b.workspaceStoreAcceptedQueryComponent {
		return
	}

	b.group.matchOnEmpty = true

	g := ids.MakeGenre()
	dq, ok := b.group.userQueries[g]

	if ok {
		delete(b.group.userQueries, g)
	} else {
		dq = b.makeQuery()
	}

	dq.Genre = b.defaultGenres

	if b.defaultSigil.IsEmpty() {
		dq.Sigil = ids.SigilLatest
	} else {
		dq.Sigil = b.defaultSigil
	}

	b.group.userQueries[b.defaultGenres] = dq
}

func (state *buildState) parseTokens() (err error) {
	q := state.makeQuery()
	stack := []stackEl{q}

	isNegated := false
	isExact := false

LOOP:
	for state.scanner.Scan() {
		seq := state.scanner.GetSeq()

		if seq.MatchAll(box.TokenTypeOperator) {
			op := seq.At(0).Contents[0]

			switch op {
			case '=':
				isExact = true

			case '^':
				isNegated = true

			case ' ':
				if len(stack) == 1 {
					break LOOP
				}

			case ',':
				last := stack[len(stack)-1].(*expTagsOrTypes)
				last.Or = true
				// TODO handle or when invalid

			case '[':
				exp := state.makeExp(isNegated, isExact)
				isExact = false
				isNegated = false
				stack[len(stack)-1].Add(exp)
				stack = append(stack, exp)

			case ']':
				stack = stack[:len(stack)-1]
				// TODO handle errors of unbalanced

			case '.':
				// TODO end sigil or embedded as part of name
				fallthrough

			case ':', '+', '?':
				if len(stack) > 1 {
					err = errors.ErrorWithStackf("sigil before end")
					return
				}

				state.scanner.Unscan()

				if err = state.parseSigilsAndGenres(q); err != nil {
					err = errors.Wrapf(err, "Seq: %q", seq)
					return
				}

				continue LOOP
			}
		} else {
			if ok, left, right, partition := seq.PartitionFavoringRight(
				box.TokenMatcherOp(box.OpSigilExternal),
			); ok {
				switch {

				// left: one/uno, partition: ., right: zettel
				case right.MatchAll(box.TokenTypeIdentifier):
					if err = q.AddString(string(right.At(0).Contents)); err != nil {
						err = nil
					} else {
						if err = state.addSigilFromOp(q, partition.Contents[0]); err != nil {
							err = errors.Wrap(err)
							return
						}

						seq = left
					}

					// left: !md, partition: ., right: ''
				case right.Len() == 0:
					if err = state.addSigilFromOp(q, partition.Contents[0]); err != nil {
						err = nil
					} else {
						seq = left
					}
				}
			}

			objectId := ObjectId{
				ObjectId: ids.GetObjectIdPool().Get(),
			}

			// TODO if this fails, permit a workspace store to try to read this as an
			// external object ID. And if that fails, try to remove the last two
			// elements as per the above and read that and force the genre and sigils
			if err = objectId.GetObjectId().ReadFromSeq(seq); err != nil {
				err = errors.BadRequestf("not a valid object id: %q", seq)
				return
			}

			if err = objectId.reduce(state); err != nil {
				err = errors.Wrap(err)
				return
			}

			pid := pinnedObjectId{
				Sigil:    ids.SigilLatest,
				ObjectId: objectId,
			}

			switch objectId.GetGenre() {
			case genres.InventoryList, genres.Zettel, genres.Repo:
				state.pinnedObjectIds = append(
					state.pinnedObjectIds,
					pid,
				)

				if err = q.addPinnedObjectId(state, pid); err != nil {
					err = errors.Wrap(err)
					return
				}

			case genres.Tag:
				var et sku.Query

				if et, err = state.makeTagExp(&objectId); err != nil {
					err = errors.Wrap(err)
					return
				}

				exp := state.makeExp(isNegated, isExact, et)
				stack[len(stack)-1].Add(exp)

			case genres.Type:
				var t ids.Type

				if err = t.TodoSetFromObjectId(objectId.GetObjectId()); err != nil {
					err = errors.Wrap(err)
					return
				}

				if !isNegated {
					if err = state.group.types.Add(t); err != nil {
						err = errors.Wrap(err)
						return
					}
				}

				exp := state.makeExp(isNegated, isExact, &objectId)
				stack[len(stack)-1].Add(exp)
			}

			isNegated = false
			isExact = false
		}
	}

	if err = state.scanner.Error(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if q.IsEmpty() {
		return
	}

	if q.Genre.IsEmpty() && !state.builder.requireNonEmptyQuery {
		q.Genre = state.defaultGenres
	}

	if q.Sigil.IsEmpty() {
		q.Sigil = state.defaultSigil
	}

	if err = state.group.add(q); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (b *buildState) addSigilFromOp(q *expSigilAndGenre, op byte) (err error) {
	var s ids.Sigil

	if err = s.SetByte(op); err != nil {
		err = errors.Wrap(err)
		return
	}

	if !b.permittedSigil.IsEmpty() && !b.permittedSigil.ContainsOneOf(s) {
		err = errors.BadRequestf("this query cannot contain the %q sigil", s)
		return
	}

	q.Sigil.Add(s)

	return
}

func (b *buildState) parseSigilsAndGenres(
	q *expSigilAndGenre,
) (err error) {
	for b.scanner.Scan() {
		seq := b.scanner.GetSeq()

		if seq.MatchAll(box.TokenTypeOperator) {
			op := seq.At(0).Contents[0]

			switch op {
			default:
				err = errors.ErrorWithStackf("unexpected operator %q", seq)
				return

			case ' ':
				return

			case '.':
				b.group.dotOperatorActive = true
				fallthrough

			case ':', '+', '?':
				if err = b.addSigilFromOp(q, op); err != nil {
					err = errors.Wrap(err)
					return
				}
			}
		} else if seq.MatchAll(box.TokenTypeIdentifier) {
			b.scanner.Unscan()
			break
		} else {
			err = errors.ErrorWithStackf("expected operator but got %q", seq)
			return
		}
	}

	if err = b.scanner.Error(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = q.ReadFromBoxScanner(&b.scanner); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO use new generic and typed blobs
func (b *buildState) makeTagOrLuaTag(
	k *ObjectId,
) (exp sku.Query, err error) {
	exp = k

	if b.builder.objectProbeIndex == nil {
		return
	}

	sk := sku.GetTransactedPool().Get()
	defer sku.GetTransactedPool().Put(sk)

	if err = b.builder.objectProbeIndex.ReadOneObjectId(
		k,
		sk,
	); err != nil {
		if collections.IsErrNotFound(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	var twb sku.TransactedWithBlob[tag_blobs.Blob]

	if twb, _, err = b.builder.typedBlobStore.Tag.GetTransactedWithBlob(
		sk,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var matcherBlob sku.Queryable

	{
		var ok bool

		if matcherBlob, ok = twb.Blob.(sku.Queryable); !ok {
			return
		}
	}

	exp = &CompoundMatch{Queryable: matcherBlob, ObjectId: k}

	return
}

func (b *buildState) makeTagExp(k *ObjectId) (exp sku.Query, err error) {
	// TODO use b.blobs to read tag blob and find filter if necessary
	var e ids.Tag

	if err = e.TodoSetFromObjectId(k.GetObjectId()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if exp, err = b.makeTagOrLuaTag(k); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (b *buildState) makeExp(
	negated, exact bool,
	children ...sku.Query,
) *expTagsOrTypes {
	return &expTagsOrTypes{
		// MatchOnEmpty: !b.doNotMatchEmpty,
		Negated:  negated,
		Exact:    exact,
		Children: children,
	}
}

func (b *buildState) makeQuery() *expSigilAndGenre {
	return &expSigilAndGenre{
		exp: exp{
			expObjectIds: expObjectIds{
				internal: make(map[string]ObjectId),
				external: make(map[string]sku.ExternalObjectId),
			},
		},
	}
}
