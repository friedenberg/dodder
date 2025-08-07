package store_config

import (
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/delta/file_extensions"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (config Config) GetFileExtensions() file_extensions.Config {
	return config.FileExtensions
}

func (compiled *compiled) getType(k interfaces.ObjectId) (ct *sku.Transacted) {
	if k.GetGenre() != genres.Type {
		return
	}

	if ct1, ok := compiled.Types.Get(k.String()); ok {
		ct = ct1.CloneTransacted()
	}

	return
}

func (compiled *compiled) getRepo(k interfaces.ObjectId) (ct *sku.Transacted) {
	if k.GetGenre() != genres.Repo {
		return
	}

	if ct1, ok := compiled.Repos.Get(k.String()); ok {
		ct = ct1.CloneTransacted()
	}

	return
}

// Returns the exactly matching Typ, or if it doesn't exist, returns the parent
// Typ or nil. (Parent Typ for `md-gdoc` would be `md`.)
func (compiled *compiled) GetApproximatedType(
	k interfaces.ObjectId,
) (ct ApproximatedType) {
	if k.GetGenre() != genres.Type {
		return
	}

	expandedActual := compiled.getSortedTypesExpanded(k.String())
	if len(expandedActual) > 0 {
		ct.HasValue = true
		ct.Type = expandedActual[0]

		if ids.Equals(ct.Type.GetObjectId(), k) {
			ct.IsActual = true
		}
	}

	return
}

func (compiled *compiled) GetTagOrRepoIdOrType(
	v string,
) (sk *sku.Transacted, err error) {
	var k ids.ObjectId

	if err = k.Set(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	switch k.GetGenre() {
	case genres.Tag:
		sk, _ = compiled.getTag(&k)
	case genres.Repo:
		sk = compiled.getRepo(&k)
	case genres.Type:
		sk = compiled.getType(&k)

	default:
		err = genres.MakeErrUnsupportedGenre(&k)
		return
	}

	return
}

func (compiled *compiled) getTag(
	k interfaces.ObjectId,
) (ct *sku.Transacted, ok bool) {
	if k.GetGenre() != genres.Tag {
		return
	}

	v := k.String()

	compiled.lock.Lock()
	defer compiled.lock.Unlock()

	expandedMaybe := collections_value.MakeMutableValueSet[values.String](nil)
	sa := quiter.MakeFuncSetString(expandedMaybe)
	expansion.ExpanderRight.Expand(sa, v)

	var cursor *tag

	for v := range expandedMaybe.All() {
		if cursor == nil {
			cursor, _ = compiled.Tags.Get(v.String())
			continue
		}

		next, ok := compiled.Tags.Get(v.String())

		if !ok {
			continue
		}

		if len(
			next.Transacted.GetObjectId().String(),
		) > len(
			cursor.Transacted.GetObjectId().String(),
		) {
			cursor = next
		}
	}

	if cursor != nil {
		ct = sku.GetTransactedPool().Get()
		sku.Resetter.ResetWith(ct, &cursor.Transacted)
	}

	return
}

// TODO-P3 merge all the below
func (compiled *compiled) getSortedTypesExpanded(
	v string,
) (expandedActual []*sku.Transacted) {
	expandedMaybe := collections_value.MakeMutableValueSet[values.String](nil)

	sa := quiter.MakeFuncSetString(expandedMaybe)

	expansion.ExpanderRight.Expand(sa, v)
	expandedActual = make([]*sku.Transacted, 0)

	for v := range expandedMaybe.All() {
		compiled.lock.Lock()
		ct, ok := compiled.Types.Get(v.String())
		compiled.lock.Unlock()

		if ok {
			expandedActual = append(expandedActual, ct)
		}
	}

	sort.Slice(expandedActual, func(i, j int) bool {
		return len(
			expandedActual[i].GetObjectId().String(),
		) > len(
			expandedActual[j].GetObjectId().String(),
		)
	})

	return
}

func (compiled *compiled) getSortedTagsExpanded(
	v string,
) (expandedActual []*sku.Transacted) {
	compiled.lock.Lock()
	defer compiled.lock.Unlock()

	expandedMaybe := collections_value.MakeMutableValueSet[values.String](nil)
	sa := quiter.MakeFuncSetString(
		expandedMaybe,
	)
	expansion.ExpanderRight.Expand(sa, v)
	expandedActual = make([]*sku.Transacted, 0)

	for v := range expandedMaybe.All() {
		ct, ok := compiled.Tags.Get(v.String())

		if !ok {
			continue
		}

		ct1 := sku.GetTransactedPool().Get()

		sku.Resetter.ResetWith(ct1, &ct.Transacted)

		expandedActual = append(expandedActual, ct1)
	}

	sort.Slice(expandedActual, func(i, j int) bool {
		return len(
			expandedActual[i].GetObjectId().String(),
		) > len(
			expandedActual[j].GetObjectId().String(),
		)
	})

	return
}

func (compiled *compiled) GetImplicitTags(
	e *ids.Tag,
) ids.TagSet {
	s, ok := compiled.ImplicitTags[e.String()]

	if !ok || s == nil {
		return ids.MakeTagSet()
	}

	return s
}
