package queries

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/doddish"
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
	latentErrors errors.GroupBuilder
	missingBlobs []ErrBlobMissing

	luaVMPoolBuilder        *lua.VMPoolBuilder
	pinnedObjectIds         []pinnedObjectId
	pinnedExternalObjectIds []sku.ExternalObjectId
	workspaceStore          store_workspace.Store

	workspaceStoreAcceptedQueryComponent bool

	scanner doddish.Scanner
}

func (src *buildState) copy() (dst *buildState) {
	dst = &buildState{
		options: src.options,
		builder: src.builder,
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

	return dst
}

func (buildState *buildState) makeGroup() *Query {
	return &Query{
		hidden:           buildState.builder.hidden,
		optimizedQueries: make(map[genres.Genre]*expSigilAndGenre),
		userQueries:      make(map[ids.Genre]*expSigilAndGenre),
		types:            ids.MakeMutableTypeSet(),
	}
}

func (buildState *buildState) build(
	values ...string,
) (err error, latent *errors.GroupBuilder) {
	latent = errors.MakeGroupBuilder()

	var remaining []string

	if buildState.workspaceStore == nil {
		remaining = values
	} else {
		for _, value := range values {
			if value == "." {
				buildState.group.dotOperatorActive = true
				remaining = append(remaining, value)
			}

			var externalObjectIds []sku.ExternalObjectId

			if externalObjectIds, err = buildState.workspaceStore.GetObjectIdsForString(
				value,
			); err != nil {
				if value != "." {
					remaining = append(remaining, value)
				}

				latent.Add(err)
				err = nil

				continue
			}

			buildState.workspaceStoreAcceptedQueryComponent = true

			for _, externalObjectId := range externalObjectIds {
				if externalObjectId.GetGenre() == genres.None {
					err = errors.ErrorWithStackf("id with empty genre: %q", externalObjectId)
					return err, latent
				}

				buildState.pinnedExternalObjectIds = append(
					buildState.pinnedExternalObjectIds,
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
	buildState.scanner.Reset(reader)

	for buildState.scanner.CanScan() {
		if err = buildState.parseTokens(); err != nil {
			err = errors.Wrap(err)
			return err, latent
		}
	}

	for _, k := range buildState.pinnedExternalObjectIds {
		if k.GetGenre() == genres.None {
			err = errors.ErrorWithStackf("id with empty genre: %q", k)
			return err, latent
		}

		if err = buildState.group.addExactExternalObjectId(buildState, k); err != nil {
			err = errors.Wrap(err)
			return err, latent
		}
	}

	for _, k := range buildState.pinnedObjectIds {
		q := buildState.makeQuery()

		if err = q.addPinnedObjectId(buildState, k); err != nil {
			err = errors.Wrap(err)
			return err, latent
		}

		if err = buildState.group.add(q); err != nil {
			err = errors.Wrap(err)
			return err, latent
		}
	}

	buildState.addDefaultsIfNecessary()

	if err = buildState.group.reduce(buildState); err != nil {
		err = errors.Wrap(err)
		return err, latent
	}

	return err, latent
}

func (buildState *buildState) addDefaultsIfNecessary() {
	if buildState.defaultGenres.IsEmpty() || !buildState.group.isEmpty() {
		return
	}

	if buildState.builder.requireNonEmptyQuery && buildState.group.isEmpty() {
		return
	}

	if buildState.workspaceStoreAcceptedQueryComponent {
		return
	}

	buildState.group.matchOnEmpty = true

	g := ids.MakeGenre()
	dq, ok := buildState.group.userQueries[g]

	if ok {
		delete(buildState.group.userQueries, g)
	} else {
		dq = buildState.makeQuery()
	}

	dq.Genre = buildState.defaultGenres

	if buildState.defaultSigil.IsEmpty() {
		dq.Sigil = ids.SigilLatest
	} else {
		dq.Sigil = buildState.defaultSigil
	}

	buildState.group.userQueries[buildState.defaultGenres] = dq
}

func (buildState *buildState) parseTokens() (err error) {
	query := buildState.makeQuery()
	stack := []stackEl{query}

	isNegated := false
	isExact := false

LOOP:
	for buildState.scanner.Scan() {
		seq := buildState.scanner.GetSeq()

		if seq.MatchAll(doddish.TokenTypeOperator) {
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
				exp := buildState.makeExp(isNegated, isExact)
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
					return err
				}

				buildState.scanner.Unscan()

				if err = buildState.parseSigilsAndGenres(query); err != nil {
					err = errors.Wrapf(err, "Seq: %q", seq)
					return err
				}

				continue LOOP
			}
		} else {
			if ok, left, right, partition := seq.PartitionFavoringRight(
				doddish.TokenMatcherOp(doddish.OpSigilExternal),
			); ok {
				switch {

				// left: one/uno, partition: ., right: zettel
				case right.MatchAll(doddish.TokenTypeIdentifier):
					if err = query.AddString(string(right.At(0).Contents)); err != nil {
						err = nil
					} else {
						if err = buildState.addSigilFromOp(query, partition.Contents[0]); err != nil {
							err = errors.Wrap(err)
							return err
						}

						seq = left
					}

					// left: !md, partition: ., right: ''
				case right.Len() == 0:
					if err = buildState.addSigilFromOp(query, partition.Contents[0]); err != nil {
						err = nil
					} else {
						seq = left
					}
				}
			}

			objectId := ObjectId{
				ObjectId: ids.GetObjectIdPool().Get(),
			}

			// TODO if this fails, permit a workspace store to try to read this
			// as an
			// external object ID. And if that fails, try to remove the last two
			// elements as per the above and read that and force the genre and
			// sigils
			if err = objectId.ReadFromSeq(seq); err != nil {
				err = errors.Wrap(err)
				return err
			}

			if err = objectId.reduce(buildState); err != nil {
				err = errors.Wrap(err)
				return err
			}

			pid := pinnedObjectId{
				Sigil:    ids.SigilLatest,
				ObjectId: objectId,
			}

			switch objectId.GetGenre() {
			case genres.InventoryList, genres.Zettel, genres.Repo:
				buildState.pinnedObjectIds = append(
					buildState.pinnedObjectIds,
					pid,
				)

				if err = query.addPinnedObjectId(buildState, pid); err != nil {
					err = errors.Wrap(err)
					return err
				}

			case genres.Tag:
				var et sku.Query

				if et, err = buildState.makeTagExp(&objectId); err != nil {
					err = errors.Wrap(err)
					return err
				}

				exp := buildState.makeExp(isNegated, isExact, et)
				stack[len(stack)-1].Add(exp)

			case genres.Type:
				var t ids.Type

				if err = t.TodoSetFromObjectId(objectId.GetObjectId()); err != nil {
					err = errors.Wrap(err)
					return err
				}

				if !isNegated {
					if err = buildState.group.types.Add(t); err != nil {
						err = errors.Wrap(err)
						return err
					}
				}

				exp := buildState.makeExp(isNegated, isExact, &objectId)
				stack[len(stack)-1].Add(exp)
			}

			isNegated = false
			isExact = false
		}
	}

	if err = buildState.scanner.Error(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if query.IsEmpty() {
		return err
	}

	if query.Genre.IsEmpty() && !buildState.builder.requireNonEmptyQuery {
		query.Genre = buildState.defaultGenres
	}

	if query.Sigil.IsEmpty() {
		query.Sigil = buildState.defaultSigil
	}

	if err = buildState.group.add(query); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (buildState *buildState) addSigilFromOp(
	q *expSigilAndGenre,
	op byte,
) (err error) {
	var s ids.Sigil

	if err = s.SetByte(op); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if !buildState.permittedSigil.IsEmpty() &&
		!buildState.permittedSigil.ContainsOneOf(s) {
		err = errors.BadRequestf("this query cannot contain the %q sigil", s)
		return err
	}

	q.Sigil.Add(s)

	return err
}

func (buildState *buildState) parseSigilsAndGenres(
	q *expSigilAndGenre,
) (err error) {
	for buildState.scanner.Scan() {
		seq := buildState.scanner.GetSeq()

		if seq.MatchAll(doddish.TokenTypeOperator) {
			op := seq.At(0).Contents[0]

			switch op {
			default:
				err = errors.ErrorWithStackf("unexpected operator %q", seq)
				return err

			case ' ':
				return err

			case '.':
				buildState.group.dotOperatorActive = true
				fallthrough

			case ':', '+', '?':
				if err = buildState.addSigilFromOp(q, op); err != nil {
					err = errors.Wrap(err)
					return err
				}
			}
		} else if seq.MatchAll(doddish.TokenTypeIdentifier) {
			buildState.scanner.Unscan()
			break
		} else {
			err = errors.ErrorWithStackf("expected operator but got %q", seq)
			return err
		}
	}

	if err = buildState.scanner.Error(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = q.ReadFromBoxScanner(&buildState.scanner); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO use new generic and typed blobs
func (buildState *buildState) makeTagOrLuaTag(
	objectId *ObjectId,
) (exp sku.Query, err error) {
	exp = objectId

	if buildState.builder.objectProbeIndex == nil {
		return exp, err
	}

	object := sku.GetTransactedPool().Get()
	defer sku.GetTransactedPool().Put(object)

	if err = buildState.builder.objectProbeIndex.ReadOneObjectId(
		objectId,
		object,
	); err != nil {
		if collections.IsErrNotFound(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return exp, err
	}

	var tagBlob tag_blobs.Blob

	if tagBlob, _, err = buildState.builder.typedBlobStore.Tag.GetBlob(
		object,
	); err != nil {
		err = errors.Wrap(err)
		return exp, err
	}

	var matcherBlob sku.Queryable

	{
		var ok bool

		if matcherBlob, ok = tagBlob.(sku.Queryable); !ok {
			return exp, err
		}
	}

	exp = &CompoundMatch{Queryable: matcherBlob, ObjectId: objectId}

	return exp, err
}

func (buildState *buildState) makeTagExp(
	k *ObjectId,
) (exp sku.Query, err error) {
	// TODO use b.blobs to read tag blob and find filter if necessary
	var e ids.Tag

	if err = e.TodoSetFromObjectId(k.GetObjectId()); err != nil {
		err = errors.Wrap(err)
		return exp, err
	}

	if exp, err = buildState.makeTagOrLuaTag(k); err != nil {
		err = errors.Wrap(err)
		return exp, err
	}

	return exp, err
}

func (buildState *buildState) makeExp(
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

func (buildState *buildState) makeQuery() *expSigilAndGenre {
	return &expSigilAndGenre{
		exp: exp{
			expObjectIds: expObjectIds{
				internal: make(map[string]ObjectId),
				external: make(map[string]sku.ExternalObjectId),
			},
		},
	}
}
