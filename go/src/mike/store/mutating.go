package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/file_lock"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/object_id_provider"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

// Saves the blob if necessary, applies the proto object, runs pre-commit hooks,
// runs the new hook, validates the blob, then calculates the digest for the
// object
func (store *Store) tryPrecommit(
	external sku.ExternalLike,
	mutter *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	if err = store.SaveBlob(external); err != nil {
		err = errors.Wrap(err)
		return
	}

	object := external.GetSku()

	if mutter == nil {
		options.Proto.Apply(object, object)
	}

	// TODO decide if the type proto should actually be applied every time
	if options.ApplyProtoType {
		store.protoZettel.ApplyType(object, object)
	}

	if genres.Type == external.GetSku().GetGenre() {
		if external.GetSku().GetType().IsEmpty() {
			external.GetSku().GetMetadata().Type = ids.DefaultOrPanic(
				genres.Type,
			)
		}
	}

	// modify pre commit hooks to support import
	if err = store.tryPreCommitHooks(object, mutter, options); err != nil {
		if store.storeConfig.GetConfig().IgnoreHookErrors {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	// TODO just just mutter == nil
	if mutter == nil {
		if err = store.tryNewHook(object, options); err != nil {
			if store.storeConfig.GetConfig().IgnoreHookErrors {
				err = nil
			} else {
				err = errors.Wrap(err)
				return
			}
		}
	}

	if err = store.validate(object, mutter, options); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = object.CalculateObjectDigestSelfWithTai(
		object_inventory_format.GetDigestForContext,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO add RealizeAndOrStore result
// TODO switch to using a child context for each object commit
func (store *Store) Commit(
	external sku.ExternalLike,
	options sku.CommitOptions,
) (err error) {
	child := external.GetSku()

	ui.Log().Printf("%s -> %s", options, child)

	// TODO remove this lock check and perform it when actually necessary (when
	// persisting the changes on flush).
	if !store.GetEnvRepo().GetLockSmith().IsAcquired() &&
		(options.AddToInventoryList || options.StreamIndexOptions.AddToStreamIndex) {
		err = errors.Wrap(file_lock.ErrLockRequired{
			Operation: "commit",
		})

		return
	}

	// TAI must be set before calculating object sha
	if options.UpdateTai {
		if options.Clock == nil {
			options.Clock = store
		}

		child.SetTai(options.Clock.GetTai())
	}

	if options.AddToInventoryList && (child.ObjectId.IsEmpty() ||
		child.GetGenre() == genres.None ||
		child.GetGenre() == genres.Blob) {
		var zettelId *ids.ZettelId

		if zettelId, err = store.zettelIdIndex.CreateZettelId(); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = child.ObjectId.SetWithIdLike(zettelId); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	var mother *sku.Transacted

	if mother, err = store.fetchMotherIfNecessary(child); err != nil {
		err = errors.Wrap(err)
		return
	}

	if mother != nil {
		defer sku.GetTransactedPool().Put(mother)
		child.Metadata.Cache.ParentTai = mother.GetTai()
	}

	if err = store.tryPrecommit(external, mother, options); err != nil {
		err = errors.Wrap(err)
		return
	}

	{
		if options.AddToInventoryList {
			if err = store.addMissingTypeAndTags(options, child); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		if options.AddToInventoryList ||
			options.StreamIndexOptions.AddToStreamIndex {
			if err = store.addObjectToAbbrStore(child); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		// short circuits if the parent is equal to the child
		if options.AddToInventoryList &&
			mother != nil &&
			ids.Equals(child.GetObjectId(), mother.GetObjectId()) &&
			child.Metadata.EqualsSansTai(&mother.Metadata) {

			sku.TransactedResetter.ResetWithExceptFields(child, mother)

			if store.sunrise.Less(child.GetTai()) {
				if err = store.ui.TransactedUnchanged(child); err != nil {
					err = errors.Wrap(err)
					return
				}
			}

			return
		}

		if err = store.applyDormantAndRealizeTags(
			child,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if child.GetGenre() == genres.Zettel {
			if err = store.zettelIdIndex.AddZettelId(&child.ObjectId); err != nil {
				if errors.Is(err, object_id_provider.ErrDoesNotExist{}) {
					ui.Log().Printf("object id does not contain value: %s", err)
					err = nil
				} else {
					err = errors.Wrapf(err, "failed to write zettel to index: %s", child)
					return
				}
			}
		}
	}

	if options.AddToInventoryList {
		// external.GetSku().Metadata.GetObjectSigMutable().Reset()
		external.GetSku().Metadata.GetObjectDigestMutable().Reset()

		if err = store.commitTransacted(child, mother); err != nil {
			err = errors.Wrapf(err, "Sku: %s", sku.String(child))
			return
		}
	}

	if options.AddToInventoryList ||
		options.StreamIndexOptions.AddToStreamIndex {
		if err = merkle_ids.MakeErrIsNull(
			child.Metadata.GetObjectSig(),
			"object-sig",
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = store.storeConfig.AddTransacted(
			child,
			mother,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = store.GetStreamIndex().Add(
			child,
			options,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if mother == nil {
			if child.GetGenre() == genres.Zettel {
				// TODO if this is a local zettel (i.e., not a different repo
				// and not a
				// different domain)

				// TODO verify that the zettel id consists of our identifiers,
				// otherwise
				// abort
			}

			if err = store.ui.TransactedNew(child); err != nil {
				err = errors.Wrap(err)
				return
			}
		} else {
			// [are/kabuto !task project-2021-zit-features zz-inbox] add delta
			// printing to changed objects
			// if err = s.Updated(mutter); err != nil {
			// 	err = errors.Wrap(err)
			// 	return
			// }

			if err = store.ui.TransactedUpdated(child); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

	}

	if options.MergeCheckedOut {
		if err = store.ReadExternalAndMergeIfNecessary(
			child,
			mother,
			options,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (store *Store) fetchMotherIfNecessary(
	object *sku.Transacted,
) (mother *sku.Transacted, err error) {
	mother = sku.GetTransactedPool().Get()
	// TODO find a way to make this more performant when operating over sshfs
	if err = store.GetStreamIndex().ReadOneObjectId(
		object.GetObjectId(),
		mother,
	); err != nil {
		if collections.IsErrNotFound(err) || errors.IsNotExist(err) {
			// TODO decide if this should continue to virtual stores
			sku.GetTransactedPool().Put(mother)
			mother = nil
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	if err = object.SetMother(mother); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO add results for which stores had which change types
func (store *Store) commitTransacted(
	object *sku.Transacted,
	mother *sku.Transacted,
) (err error) {
	if !store.inventoryList.LastTai.Less(object.GetTai()) {
		object.Metadata.Tai = store.GetTai()
	}

	if err = store.inventoryListStore.AddObjectToOpenList(
		store.inventoryList,
		object,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) handleUnchanged(
	t *sku.Transacted,
) (err error) {
	return store.ui.TransactedUnchanged(t)
}

func (store *Store) UpdateKonfig(
	sh interfaces.BlobId,
) (kt *sku.Transacted, err error) {
	return store.CreateOrUpdateBlobSha(
		&ids.Config{},
		sh,
	)
}

func (store *Store) createTagsOrType(k *ids.ObjectId) (err error) {
	t := sku.GetTransactedPool().Get()
	defer sku.GetTransactedPool().Put(t)

	switch k.GetGenre() {
	default:
		err = genres.MakeErrUnsupportedGenre(k.GetGenre())
		return

	case genres.Type:
		t.GetMetadata().Type = ids.DefaultOrPanic(genres.Type)

	case genres.Tag:
	}

	if err = t.ObjectId.SetWithIdLike(k); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.addObjectToAbbrStore(t); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.Commit(
		t,
		sku.CommitOptions{
			StoreOptions:       sku.GetStoreOptionsUpdate(),
			DontAddMissingTags: true,
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) addType(
	t ids.Type,
) (err error) {
	if t.IsEmpty() {
		err = errors.ErrorWithStackf("attempting to add empty type")
		return
	}

	var oid ids.ObjectId

	if err = oid.ResetWithIdLike(t); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.GetStreamIndex().ObjectExists(&oid); err == nil {
		return
	}

	err = nil

	var k ids.ObjectId

	if err = k.SetWithIdLike(t); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.createTagsOrType(&k); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) addTypeAndExpandedIfNecessary(
	rootTipe ids.Type,
) (err error) {
	if rootTipe.IsEmpty() {
		return
	}

	if ids.IsBuiltin(rootTipe) {
		return
	}

	typesExpanded := ids.ExpandOneSlice(
		rootTipe,
		ids.MakeType,
		expansion.ExpanderRight,
	)

	for _, tipe := range typesExpanded {
		if err = store.addType(tipe); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (store *Store) addTags(
	tags []ids.Tag,
) (err error) {
	store.tagLock.Lock()
	defer store.tagLock.Unlock()

	var oid ids.ObjectId

	for _, tag := range tags {
		if tag.IsVirtual() {
			continue
		}

		if err = oid.ResetWithIdLike(tag); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = store.GetStreamIndex().ObjectExists(&oid); err == nil {
			continue
		}

		err = nil

		var k ids.ObjectId

		if err = k.SetWithIdLike(tag); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = store.createTagsOrType(&k); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (store *Store) addMissingTypeAndTags(
	commitOptions sku.CommitOptions,
	object *sku.Transacted,
) (err error) {
	tipe := object.GetType()

	if !commitOptions.DontAddMissingType {
		if err = store.addTypeAndExpandedIfNecessary(tipe); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if !commitOptions.DontAddMissingTags && object.GetGenre() == genres.Tag {
		var tag ids.Tag

		if err = tag.TodoSetFromObjectId(object.GetObjectId()); err != nil {
			err = errors.Wrap(err)
			return
		}

		tagsExpanded := ids.ExpandOneSlice(
			tag,
			ids.MakeTag,
			expansion.ExpanderRight,
		)

		if len(tagsExpanded) > 0 {
			tagsExpanded = tagsExpanded[:len(tagsExpanded)-1]
		}

		if err = store.addTags(tagsExpanded); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if !commitOptions.DontAddMissingTags {
		es := quiter.SortedValues(object.Metadata.GetTags())

		if object.GetGenre() == genres.Tag {
			var tag ids.Tag

			if err = tag.TodoSetFromObjectId(object.GetObjectId()); err != nil {
				err = errors.Wrap(err)
				return
			}

			tagsExpanded := ids.ExpandOneSlice(
				tag,
				ids.MakeTag,
				expansion.ExpanderRight,
			)

			if len(tagsExpanded) > 0 {
				tagsExpanded = tagsExpanded[:len(tagsExpanded)-1]
			}

			if err = store.addTags(tagsExpanded); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		for _, e := range es {
			tagsExpanded := ids.ExpandOneSlice(
				e,
				ids.MakeTag,
				expansion.ExpanderRight,
			)

			if err = store.addTags(tagsExpanded); err != nil {
				err = errors.Wrap(err)
				return
			}
		}
	}

	return
}

func (store *Store) addObjectToAbbrStore(m *sku.Transacted) (err error) {
	if err = store.GetAbbrStore().AddObjectToAbbreviationStore(m); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) reindexOne(object sku.ObjectWithList) (err error) {
	o := sku.CommitOptions{
		StoreOptions: sku.GetStoreOptionsReindex(),
	}

	if err = store.Commit(object.Object, o); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.GetAbbrStore().AddObjectToAbbreviationStore(
		object.Object,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
