package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/tag_paths"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

// TODO extract into store_tags
func (store *Store) applyDormantAndRealizeTags(
	object *sku.Transacted,
) (err error) {
	ui.Log().Print("applying konfig to:", object)
	mp := &object.Metadata

	mp.Cache.SetExpandedTags(ids.ExpandMany(
		mp.GetTags(),
		expansion.ExpanderRight,
	))

	g := genres.Must(object.GetGenre())
	isTag := g == genres.Tag

	// if g.HasParents() {
	// 	k.SetHasChanges(fmt.Sprintf("adding etikett with parents: %s", sk))
	// }

	var tag ids.Tag

	// TODO better solution for "realizing" tags against Config.
	// Specifically, making this less fragile and dependent on remembering to do
	// ApplyToSku for each Sku. Maybe a factory?
	mp.Cache.TagPaths.Reset()
	mp.GetTags().Each(mp.Cache.TagPaths.AddTagOld)

	if isTag {
		ks := object.ObjectId.String()

		if err = tag.Set(ks); err != nil {
			err = errors.Wrap(err)
			return
		}

		object.Metadata.Cache.TagPaths.AddSelf(catgut.MakeFromString(ks))

		ids.ExpandOneInto(
			tag,
			ids.MakeTag,
			expansion.ExpanderRight,
			mp.Cache.GetExpandedTagsMutable(),
		)
	}

	if err = store.addSuperTags(object); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.addImplicitTags(object); err != nil {
		err = errors.Wrap(err)
		return
	}

	object.SetDormant(store.dormantIndex.ContainsSku(object))

	return
}

func (store *Store) addSuperTags(
	sk *sku.Transacted,
) (err error) {
	g := sk.GetGenre()

	var expanded []string
	var ks string

	switch g {
	case genres.Tag, genres.Type, genres.Repo:
		ks = sk.ObjectId.String()

		expansion.ExpanderRight.Expand(
			func(v string) (err error) {
				expanded = append(expanded, v)
				return
			},
			ks,
		)

	default:
		return
	}

	for _, ex := range expanded {
		if ex == ks || ex == "" {
			continue
		}

		func() {
			var ek *sku.Transacted

			if ek, err = store.storeConfig.GetConfig().GetTagOrRepoIdOrType(ex); err != nil {
				err = errors.Wrapf(err, "Expanded: %q", ex)
				return
			}

			if ek == nil {
				// this is ok because currently, konfig is applied twice.
				// However, this
				// is fragile as the order in which this method is called is
				// non-deterministic and the `GetTag` call may request an Tag we
				// have not processed yet
				return
			}

			defer sku.GetTransactedPool().Put(ek)

			if ek.Metadata.Cache.TagPaths.Paths.Len() <= 1 {
				ui.Log().Print(ks, ex, ek.Metadata.Cache.TagPaths)
				return
			}

			prefix := catgut.MakeFromString(ex)

			a := &sk.Metadata.Cache.TagPaths
			b := &ek.Metadata.Cache.TagPaths

			ui.Log().Print("a", a)
			ui.Log().Print("b", b)

			ui.Log().Print("prefix", prefix)

			if err = a.AddSuperFrom(b, prefix); err != nil {
				err = errors.Wrap(err)
				return
			}

			ui.Log().Print("a after", a)
		}()
	}

	return
}

func (store *Store) addImplicitTags(
	sk *sku.Transacted,
) (err error) {
	mp := &sk.Metadata
	ie := ids.MakeTagMutableSet()

	addImplicitTags := func(e *ids.Tag) (err error) {
		p1 := tag_paths.MakePathWithType()
		p1.Type = tag_paths.TypeIndirect
		p1.Add(catgut.MakeFromString(e.String()))

		implicitTags := store.storeConfig.GetConfig().GetImplicitTags(e)

		if implicitTags.Len() == 0 {
			sk.Metadata.Cache.TagPaths.AddPathWithType(p1)
			return
		}

		for e1 := range implicitTags.All() {
			p2 := p1.Clone()
			p2.Add(catgut.MakeFromString(e1.String()))
			sk.Metadata.Cache.TagPaths.AddPathWithType(p2)
		}

		return
	}

	mp.GetTags().EachPtr(addImplicitTags)

	typKonfig := store.storeConfig.GetConfig().GetApproximatedType(
		mp.GetType(),
	).ApproximatedOrActual()

	if typKonfig != nil {
		typKonfig.GetTags().EachPtr(ie.AddPtr)
		typKonfig.GetTags().EachPtr(addImplicitTags)
	}

	mp.Cache.SetImplicitTags(ie)

	return
}
