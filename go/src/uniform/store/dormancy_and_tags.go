package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/echo/catgut"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/golf/tag_paths"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

// TODO extract into store_tags
func (store *Store) applyDormantAndRealizeTags(
	object *sku.Transacted,
) (err error) {
	ui.Log().Print("applying konfig to:", object)
	metadata := object.GetMetadataMutable()

	genre := genres.Must(object.GetGenre())
	isTag := genre == genres.Tag

	// if g.HasParents() {
	// 	k.SetHasChanges(fmt.Sprintf("adding etikett with parents: %s", sk))
	// }

	var tag ids.Tag

	// TODO better solution for "realizing" tags against Config.
	// Specifically, making this less fragile and dependent on remembering to do
	// ApplyToSku for each Sku. Maybe a factory?
	metadata.GetIndexMutable().GetTagPathsMutable().Reset()
	for tag := range metadata.AllTags() {
		metadata.GetIndexMutable().GetTagPathsMutable().AddTagOld(tag)
	}

	if isTag {
		objectIdString := object.ObjectId.String()

		if err = tag.Set(objectIdString); err != nil {
			err = errors.Wrap(err)
			return err
		}

		object.GetMetadataMutable().GetIndexMutable().GetTagPaths().AddSelf(
			catgut.MakeFromString(objectIdString),
		)
	}

	if err = store.addSuperTags(object); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = store.addImplicitTags(object); err != nil {
		err = errors.Wrap(err)
		return err
	}

	object.SetDormant(store.dormantIndex.ContainsSku(object))

	return err
}

func (store *Store) addSuperTags(
	object *sku.Transacted,
) (err error) {
	genre := object.GetGenre()

	var expanded collections_slice.Slice[string]
	var objectIdString string

	switch genre {
	case genres.Tag, genres.Type, genres.Repo:
		objectIdString = object.ObjectId.String()

		quiter.AppendSeq(
			&expanded,
			expansion.ExpanderRight.Expand(objectIdString),
		)

	default:
		return err
	}

	for _, expandedObjectIdComponent := range expanded {
		if expandedObjectIdComponent == objectIdString ||
			expandedObjectIdComponent == "" {
			continue
		}

		func() {
			var tagOrRepoOrTypeObject *sku.Transacted

			if tagOrRepoOrTypeObject, err = store.storeConfig.GetConfig().GetTagOrRepoIdOrType(
				expandedObjectIdComponent,
			); err != nil {
				err = errors.Wrapf(
					err,
					"Expanded: %q",
					expandedObjectIdComponent,
				)
				return
			}

			if tagOrRepoOrTypeObject == nil {
				// this is ok because currently, konfig is applied twice.
				// However, this
				// is fragile as the order in which this method is called is
				// non-deterministic and the `GetTag` call may request an Tag we
				// have not processed yet
				return
			}

			defer sku.GetTransactedPool().Put(tagOrRepoOrTypeObject)

			if tagOrRepoOrTypeObject.GetMetadata().GetIndex().GetTagPaths().Paths.Len() <= 1 {
				ui.Log().Print(
					objectIdString,
					expandedObjectIdComponent,
					tagOrRepoOrTypeObject.GetMetadata().GetIndex().GetTagPaths(),
				)
				return
			}

			prefix := catgut.MakeFromString(expandedObjectIdComponent)

			a := object.GetMetadataMutable().GetIndexMutable().GetTagPaths()
			b := tagOrRepoOrTypeObject.GetMetadata().GetIndex().GetTagPaths()

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

	return err
}

func (store *Store) addImplicitTags(
	object *sku.Transacted,
) (err error) {
	metadata := object.GetMetadataMutable()
	tagSet := ids.MakeTagSetMutable()

	addImplicitTags := func(tag ids.Tag) (err error) {
		tagPathWithType := tag_paths.MakePathWithType()
		tagPathWithType.Type = tag_paths.TypeIndirect
		tagPathWithType.Add(catgut.MakeFromString(tag.String()))

		implicitTags := store.storeConfig.GetConfig().GetImplicitTags(tag)

		if implicitTags.Len() == 0 {
			object.GetMetadataMutable().GetIndexMutable().GetTagPaths().AddPathWithType(tagPathWithType)
			return err
		}

		for implicitTag := range implicitTags.All() {
			tagPathWithTypeClone := tagPathWithType.Clone()
			tagPathWithTypeClone.Add(
				catgut.MakeFromString(implicitTag.String()),
			)
			object.GetMetadataMutable().GetIndexMutable().GetTagPaths().AddPathWithType(tagPathWithTypeClone)
		}

		return err
	}

	for tag := range metadata.AllTags() {
		if err = addImplicitTags(tag); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	typeObject := store.storeConfig.GetConfig().GetApproximatedType(
		metadata.GetType(),
	).ApproximatedOrActual()

	if typeObject != nil {
		for tag := range typeObject.GetMetadata().AllTags() {
			if err = tagSet.Add(tag); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}

		for tag := range typeObject.GetMetadata().AllTags() {
			if err = addImplicitTags(tag); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}
	}

	metadata.GetIndexMutable().SetImplicitTags(tagSet)

	return err
}
