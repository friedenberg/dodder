package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
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
	mother *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	if err = store.SaveBlob(external); err != nil {
		err = errors.Wrap(err)
		return
	}

	object := external.GetSku()

	if mother == nil {
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
	if err = store.tryPreCommitHooks(object, mother, options); err != nil {
		if store.storeConfig.GetConfig().IgnoreHookErrors {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	// TODO just just mutter == nil
	if mother == nil {
		if err = store.tryNewHook(object, options); err != nil {
			if store.storeConfig.GetConfig().IgnoreHookErrors {
				err = nil
			} else {
				err = errors.Wrap(err)
				return
			}
		}
	}

	if err = store.validate(object, mother, options); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = object.CalculateObjectDigestSelfWithoutTai(
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
	daughter := external.GetSku()

	ui.Log().Printf("%s -> %s", options, daughter)

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

		tai := options.Clock.GetTai()
		daughter.SetTai(tai)
	}

	if options.AddToInventoryList && (daughter.ObjectId.IsEmpty() ||
		daughter.GetGenre() == genres.None ||
		daughter.GetGenre() == genres.Blob) {
		var zettelId *ids.ZettelId

		if zettelId, err = store.zettelIdIndex.CreateZettelId(); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = daughter.ObjectId.SetWithIdLike(zettelId); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	var mother *sku.Transacted

	if mother, err = store.fetchMotherIfNecessary(daughter); err != nil {
		err = errors.Wrap(err)
		return
	}

	if mother != nil {
		defer sku.GetTransactedPool().Put(mother)
		daughter.Metadata.Cache.ParentTai = mother.GetTai()
	}

	if err = store.tryPrecommit(external, mother, options); err != nil {
		err = errors.Wrap(err)
		return
	}

	{
		if options.AddToInventoryList {
			if err = store.addMissingTypeAndTags(options, daughter); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		if options.AddToInventoryList ||
			options.StreamIndexOptions.AddToStreamIndex {
			if err = store.addObjectToAbbrStore(daughter); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		// short circuits if the parent is equal to the child
		if options.AddToInventoryList &&
			mother != nil &&
			ids.Equals(daughter.GetObjectId(), mother.GetObjectId()) &&
			daughter.Metadata.EqualsSansTai(&mother.Metadata) {

			sku.TransactedResetter.ResetWithExceptFields(daughter, mother)

			if store.sunrise.Less(daughter.GetTai()) {
				if err = store.ui.TransactedUnchanged(daughter); err != nil {
					err = errors.Wrap(err)
					return
				}
			}

			return
		}

		if err = store.applyDormantAndRealizeTags(
			daughter,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if daughter.GetGenre() == genres.Zettel {
			if err = store.zettelIdIndex.AddZettelId(&daughter.ObjectId); err != nil {
				if errors.Is(err, object_id_provider.ErrDoesNotExist{}) {
					ui.Log().Printf("object id does not contain value: %s", err)
					err = nil
				} else {
					err = errors.Wrapf(err, "failed to write zettel to index: %s", daughter)
					return
				}
			}
		}
	}

	if options.AddToInventoryList {
		// external.GetSku().Metadata.GetObjectSigMutable().Reset()
		external.GetSku().Metadata.GetObjectDigestMutable().Reset()

		if err = store.commitTransacted(daughter, mother); err != nil {
			err = errors.Wrapf(err, "Sku: %s", sku.String(daughter))
			return
		}
	}

	if options.AddToInventoryList ||
		options.StreamIndexOptions.AddToStreamIndex {
		if store_version.GreaterOrEqual(
			store.storeConfig.GetConfig().GetStoreVersion(),
			store_version.V11,
		) {
			if err = markl.AssertIdIsNotNull(
				daughter.Metadata.GetObjectSig(),
				"object-sig",
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		if err = store.storeConfig.AddTransacted(
			daughter,
			mother,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = store.GetStreamIndex().Add(
			daughter,
			options,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if mother == nil {
			if daughter.GetGenre() == genres.Zettel {
				// TODO if this is a local zettel (i.e., not a different repo
				// and not a
				// different domain)

				// TODO verify that the zettel id consists of our identifiers,
				// otherwise
				// abort
			}

			if err = store.ui.TransactedNew(daughter); err != nil {
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

			if err = store.ui.TransactedUpdated(daughter); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

	}

	if options.MergeCheckedOut {
		if err = store.ReadExternalAndMergeIfNecessary(
			daughter,
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
	objectId := object.GetObjectId()

	if objectId.IsEmpty() {
		return
	}

	mother = sku.GetTransactedPool().Get()
	// TODO find a way to make this more performant when operating over sshfs
	if err = store.GetStreamIndex().ReadOneObjectId(
		objectId,
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
	daughter *sku.Transacted,
	mother *sku.Transacted,
) (err error) {
	if err = store.inventoryListStore.AddObjectToOpenList(
		store.inventoryList,
		daughter,
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
	blobId interfaces.MarklId,
) (kt *sku.Transacted, err error) {
	return store.CreateOrUpdateBlobSha(
		&ids.Config{},
		blobId,
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

func (store *Store) addObjectToAbbrStore(object *sku.Transacted) (err error) {
	if err = store.GetAbbrStore().AddObjectToAbbreviationStore(
		object,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) reindexOne(object sku.ObjectWithList) (err error) {
	options := sku.CommitOptions{
		StoreOptions: sku.GetStoreOptionsReindex(),
	}

	if err = store.Commit(object.Object, options); err != nil {
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
