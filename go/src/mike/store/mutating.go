package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/file_lock"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/object_id_provider"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
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
		return err
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
			return err
		}
	}

	// TODO just just mutter == nil
	if mother == nil {
		if err = store.tryNewHook(object, options); err != nil {
			if store.storeConfig.GetConfig().IgnoreHookErrors {
				err = nil
			} else {
				err = errors.Wrap(err)
				return err
			}
		}
	}

	if err = store.validate(object, mother, options); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO add RealizeAndOrStore result
// TODO switch to using a child context for each object commit
func (store *Store) Commit(
	external sku.ExternalLike,
	options sku.CommitOptions,
) (err error) {
	daughter := external.GetSku()

	if daughter == nil {
		panic("empty daughter")
	}

	ui.Log().Printf("%s -> %s", options, daughter)

	// TODO remove this lock check and perform it when actually necessary (when
	// persisting the changes on flush).
	if !store.GetEnvRepo().GetLockSmith().IsAcquired() &&
		(options.AddToInventoryList || options.StreamIndexOptions.AddToStreamIndex) {
		err = errors.Wrap(file_lock.ErrLockRequired{
			Operation: "commit",
		})

		return err
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
			return err
		}

		if err = daughter.ObjectId.SetWithIdLike(zettelId); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	var mother *sku.Transacted

	if mother, err = store.fetchMotherIfNecessary(daughter); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if mother != nil {
		defer sku.GetTransactedPool().Put(mother)
		daughter.Metadata.Cache.ParentTai = mother.GetTai()
	}

	if err = store.tryPrecommit(external, mother, options); err != nil {
		err = errors.Wrap(err)
		return err
	}

	{
		if options.AddToInventoryList {
			if err = store.addMissingTypes(options, daughter); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}

		if options.AddToInventoryList ||
			options.StreamIndexOptions.AddToStreamIndex {
			if err = store.GetAbbrStore().AddObjectToIdIndex(
				daughter,
			); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}

		// short circuits if the parent is equal to the child
		if options.AddToInventoryList &&
			mother != nil &&
			ids.Equals(daughter.GetObjectId(), mother.GetObjectId()) &&
			daughter.Metadata.EqualsSansTai(&mother.Metadata) {

			sku.TransactedResetter.ResetWithExceptFields(daughter, mother)

			// TODO why is this condition here
			if store.sunrise.Less(daughter.GetTai()) {
				if err = store.handleUnchanged(daughter); err != nil {
					err = errors.Wrap(err)
					return err
				}
			}

			return err
		}

		if err = store.applyDormantAndRealizeTags(
			daughter,
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if daughter.GetGenre() == genres.Zettel {
			if err = store.zettelIdIndex.AddZettelId(&daughter.ObjectId); err != nil {
				if errors.Is(err, object_id_provider.ErrDoesNotExist{}) {
					ui.Log().Printf("object id does not contain value: %s", err)
					err = nil
				} else {
					err = errors.Wrapf(err, "failed to write zettel to index: %s", daughter)
					return err
				}
			}
		}
	}

	if options.AddToInventoryList {
		// external.GetSku().Metadata.GetObjectSigMutable().Reset()
		external.GetSku().Metadata.GetObjectDigestMutable().Reset()

		if err = store.commitTransacted(daughter, mother); err != nil {
			err = errors.Wrapf(err, "Sku: %s", sku.String(daughter))
			return err
		}
	}

	if options.AddToInventoryList ||
		options.StreamIndexOptions.AddToStreamIndex {
		if store_version.GreaterOrEqual(
			store.storeConfig.GetConfig().GetStoreVersion(),
			store_version.V11,
		) {
			if err = markl.AssertIdIsNotNull(
				daughter.Metadata.GetObjectSig()); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}

		if err = store.storeConfig.AddTransacted(
			daughter,
			mother,
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = store.GetStreamIndex().Add(
			daughter,
			options,
		); err != nil {
			err = errors.Wrap(err)
			return err
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
				return err
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
				return err
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
			return err
		}
	}

	return err
}

func (store *Store) fetchMotherIfNecessary(
	daughter *sku.Transacted,
) (mother *sku.Transacted, err error) {
	if daughter == nil {
		panic("empty daughter")
	}

	objectId := daughter.GetObjectId()

	if objectId.IsEmpty() {
		return mother, err
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

		return mother, err
	}

	if err = daughter.SetMother(mother); err != nil {
		err = errors.Wrap(err)
		return mother, err
	}

	return mother, err
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
		return err
	}

	return err
}

func (store *Store) handleUnchanged(
	object *sku.Transacted,
) (err error) {
	return store.ui.TransactedUnchanged(object)
}

func (store *Store) UpdateKonfig(
	blobId interfaces.MarklId,
) (kt *sku.Transacted, err error) {
	return store.CreateOrUpdateBlobDigest(
		&ids.Config{},
		blobId,
	)
}

func (store *Store) createType(typeId *ids.ObjectId) (err error) {
	typeObject := sku.GetTransactedPool().Get()
	defer sku.GetTransactedPool().Put(typeObject)

	switch typeId.GetGenre() {
	default:
		err = genres.MakeErrUnsupportedGenre(typeId.GetGenre())
		return err

	case genres.Type:
		typeObject.GetMetadata().Type = ids.DefaultOrPanic(genres.Type)
	}

	if err = typeObject.ObjectId.SetWithIdLike(typeId); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = store.Commit(
		typeObject,
		sku.CommitOptions{
			StoreOptions: sku.GetStoreOptionsUpdate(),
		},
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (store *Store) addTypeIfNecessary(
	typeId ids.Type,
) (err error) {
	if typeId.IsEmpty() {
		err = errors.ErrorWithStackf("attempting to add empty type")
		return err
	}

	var objectId ids.ObjectId

	if err = objectId.ResetWithIdLike(typeId); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = store.GetStreamIndex().ObjectExists(&objectId); err == nil {
		return err
	}

	err = nil

	var typeObjectId ids.ObjectId

	if err = typeObjectId.SetWithIdLike(typeId); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = store.createType(&typeObjectId); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (store *Store) addTypeAndExpandedIfNecessary(
	rootTipe ids.Type,
) (err error) {
	if rootTipe.IsEmpty() {
		return err
	}

	if ids.IsBuiltin(rootTipe) {
		return err
	}

	typesExpanded := ids.ExpandOneSlice(
		rootTipe,
		ids.MakeType,
		expansion.ExpanderRight,
	)

	for _, tipe := range typesExpanded {
		if err = store.addTypeIfNecessary(tipe); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (store *Store) addMissingTypes(
	commitOptions sku.CommitOptions,
	object *sku.Transacted,
) (err error) {
	tipe := object.GetType()

	if !commitOptions.DontAddMissingType {
		if err = store.addTypeAndExpandedIfNecessary(tipe); err != nil {
			err = errors.Wrap(err)
			return err
		}
	} else {
		// TODO enforce that object has signature
	}

	return err
}

func (store *Store) reindexOne(object sku.ObjectWithList) (err error) {
	options := sku.CommitOptions{
		StoreOptions: sku.GetStoreOptionsReindex(),
	}

	if err = store.Commit(object.Object, options); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = store.GetAbbrStore().AddObjectToIdIndex(
		object.Object,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
