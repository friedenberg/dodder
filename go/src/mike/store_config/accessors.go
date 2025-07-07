package store_config

import (
	"fmt"
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/config_mutable_cli"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (c *store) GetZettelFileExtension() string {
	return fmt.Sprintf(".%s", c.GetFileExtensions().GetFileExtensionZettel())
}

func (kc *store) GetImmutableConfig() config_immutable.Private {
	return kc.immutable_config_private
}

func (kc *compiled) getType(k interfaces.ObjectId) (ct *sku.Transacted) {
	if k.GetGenre() != genres.Type {
		return
	}

	if ct1, ok := kc.Types.Get(k.String()); ok {
		ct = ct1.CloneTransacted()
	}

	return
}

func (kc *compiled) getRepo(k interfaces.ObjectId) (ct *sku.Transacted) {
	if k.GetGenre() != genres.Repo {
		return
	}

	if ct1, ok := kc.Repos.Get(k.String()); ok {
		ct = ct1.CloneTransacted()
	}

	return
}

// Returns the exactly matching Typ, or if it doesn't exist, returns the parent
// Typ or nil. (Parent Typ for `md-gdoc` would be `md`.)
func (kc *compiled) GetApproximatedType(
	k interfaces.ObjectId,
) (ct ApproximatedType) {
	if k.GetGenre() != genres.Type {
		return
	}

	expandedActual := kc.getSortedTypesExpanded(k.String())
	if len(expandedActual) > 0 {
		ct.HasValue = true
		ct.Type = expandedActual[0]

		if ids.Equals(ct.Type.GetObjectId(), k) {
			ct.IsActual = true
		}
	}

	return
}

func (kc *compiled) GetTagOrRepoIdOrType(
	v string,
) (sk *sku.Transacted, err error) {
	var k ids.ObjectId

	if err = k.Set(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	switch k.GetGenre() {
	case genres.Tag:
		sk, _ = kc.getTag(&k)
	case genres.Repo:
		sk = kc.getRepo(&k)
	case genres.Type:
		sk = kc.getType(&k)

	default:
		err = genres.MakeErrUnsupportedGenre(&k)
		return
	}

	return
}

func (kc *compiled) getTag(
	k interfaces.ObjectId,
) (ct *sku.Transacted, ok bool) {
	if k.GetGenre() != genres.Tag {
		return
	}

	v := k.String()

	kc.lock.Lock()
	defer kc.lock.Unlock()

	expandedMaybe := collections_value.MakeMutableValueSet[values.String](nil)
	sa := quiter.MakeFuncSetString(expandedMaybe)
	expansion.ExpanderRight.Expand(sa, v)

	var cursor *tag

	for v := range expandedMaybe.All() {
		if cursor == nil {
			cursor, _ = kc.Tags.Get(v.String())
			continue
		}

		next, ok := kc.Tags.Get(v.String())

		if !ok {
			continue
		}

		if len(next.Transacted.GetObjectId().String()) > len(cursor.Transacted.GetObjectId().String()) {
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
func (c *compiled) getSortedTypesExpanded(
	v string,
) (expandedActual []*sku.Transacted) {
	expandedMaybe := collections_value.MakeMutableValueSet[values.String](nil)

	sa := quiter.MakeFuncSetString(expandedMaybe)

	expansion.ExpanderRight.Expand(sa, v)
	expandedActual = make([]*sku.Transacted, 0)

	expandedMaybe.Each(
		func(v values.String) (err error) {
			c.lock.Lock()
			defer c.lock.Unlock()

			ct, ok := c.Types.Get(v.String())

			if !ok {
				return
			}

			expandedActual = append(expandedActual, ct)

			return
		},
	)

	sort.Slice(expandedActual, func(i, j int) bool {
		return len(
			expandedActual[i].GetObjectId().String(),
		) > len(
			expandedActual[j].GetObjectId().String(),
		)
	})

	return
}

func (c *compiled) getSortedTagsExpanded(
	v string,
) (expandedActual []*sku.Transacted) {
	c.lock.Lock()
	defer c.lock.Unlock()

	expandedMaybe := collections_value.MakeMutableValueSet[values.String](nil)
	sa := quiter.MakeFuncSetString(
		expandedMaybe,
	)
	expansion.ExpanderRight.Expand(sa, v)
	expandedActual = make([]*sku.Transacted, 0)

	expandedMaybe.Each(
		func(v values.String) (err error) {
			ct, ok := c.Tags.Get(v.String())

			if !ok {
				return
			}

			ct1 := sku.GetTransactedPool().Get()

			sku.Resetter.ResetWith(ct1, &ct.Transacted)

			expandedActual = append(expandedActual, ct1)

			return
		},
	)

	sort.Slice(expandedActual, func(i, j int) bool {
		return len(
			expandedActual[i].GetObjectId().String(),
		) > len(
			expandedActual[j].GetObjectId().String(),
		)
	})

	return
}

func (c *compiled) GetImplicitTags(
	e *ids.Tag,
) ids.TagSet {
	s, ok := c.ImplicitTags[e.String()]

	if !ok || s == nil {
		return ids.MakeTagSet()
	}

	return s
}

func (kc *store) Cli() config_mutable_cli.Config {
	return kc.cli
}
