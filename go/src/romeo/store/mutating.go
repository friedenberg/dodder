package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/foxtrot/object_id_provider"
	"code.linenisgreat.com/dodder/go/src/india/file_lock"
	"code.linenisgreat.com/dodder/go/src/india/object_metadata"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

func (store *Store) Commit(
	external sku.ExternalLike,
	options sku.CommitOptions,
) (err error) {
	committer := commitFacilitator{
		Store: store,
		index: store.streamIndex,
	}

	if err = committer.commit(external, options); err != nil {
		err = errors.Wrapf(err, "Sku: %q", sku.String(external.GetSku()))
		return err
	}

	return err
}

type commitFacilitator struct {
	*Store
	index sku.Reindexer
}

// Saves the blob if necessary, applies the proto object, runs pre-commit hooks,
// runs the new hook, validates the blob, then calculates the digest for the
// object
func (commitFacilitator commitFacilitator) tryPrecommit(
	external sku.ExternalLike,
	mother *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	if err = commitFacilitator.SaveBlob(external); err != nil {
		err = errors.Wrap(err)
		return err
	}

	object := external.GetSku()

	if mother == nil {
		options.Proto.Apply(object, object)
	}

	// TODO decide if the type proto should actually be applied every time
	if options.ApplyProtoType {
		commitFacilitator.protoZettel.ApplyType(object, object)
	}

	if genres.Type == external.GetSku().GetGenre() {
		if external.GetSku().GetType().IsEmpty() {
			external.GetSku().GetMetadataMutable().GetTypePtr().ResetWith(
				ids.DefaultOrPanic(genres.Type),
			)
		}
	}

	// modify pre commit hooks to support import
	if err = commitFacilitator.tryPreCommitHooks(object, mother, options); err != nil {
		if commitFacilitator.storeConfig.GetConfig().IgnoreHookErrors {
			err = nil
		} else {
			err = errors.Wrap(err)
			return err
		}
	}

	// TODO just just mutter == nil
	if mother == nil {
		if err = commitFacilitator.tryNewHook(object, options); err != nil {
			if commitFacilitator.storeConfig.GetConfig().IgnoreHookErrors {
				err = nil
			} else {
				err = errors.Wrap(err)
				return err
			}
		}
	}

	if err = commitFacilitator.validateAndFinalize(
		object,
		mother,
		options,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO add RealizeAndOrStore result
// TODO switch to using a child context for each object commit
func (commitFacilitator commitFacilitator) commit(
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
	if !commitFacilitator.GetEnvRepo().GetLockSmith().IsAcquired() &&
		(options.AddToInventoryList || options.StreamIndexOptions.AddToStreamIndex) {
		err = errors.Wrap(file_lock.ErrLockRequired{
			Operation: "commit",
		})

		return err
	}

	// TAI must be set before calculating object sha
	if options.UpdateTai {
		if options.Clock == nil {
			options.Clock = commitFacilitator
		}

		tai := options.Clock.GetTai()
		daughter.SetTai(tai)
	}

	if options.AddToInventoryList && (daughter.ObjectId.IsEmpty() ||
		daughter.GetGenre() == genres.None ||
		daughter.GetGenre() == genres.Blob) {
		var zettelId *ids.ZettelId

		if zettelId, err = commitFacilitator.zettelIdIndex.CreateZettelId(); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = daughter.ObjectId.SetWithIdLike(zettelId); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	var mother *sku.Transacted

	if mother, err = commitFacilitator.fetchMotherIfNecessary(
		daughter,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if mother != nil {
		defer sku.GetTransactedPool().Put(mother)
		*daughter.GetMetadataMutable().GetIndexMutable().GetParentTaiMutable() = mother.GetTai()
	}

	if err = commitFacilitator.tryPrecommit(external, mother, options); err != nil {
		err = errors.Wrap(err)
		return err
	}

	{
		if options.AddToInventoryList {
			if err = commitFacilitator.addMissingTypes(
				options,
				daughter,
			); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}

		if options.AddToInventoryList ||
			options.StreamIndexOptions.AddToStreamIndex {
			if err = commitFacilitator.GetAbbrStore().AddObjectToIdIndex(
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
			object_metadata.EqualerSansTai.Equals(daughter.GetMetadata(), mother.GetMetadata()) {

			sku.TransactedResetter.ResetWithExceptFields(daughter, mother)

			// TODO why is this condition here
			if commitFacilitator.sunrise.Less(daughter.GetTai()) {
				if err = commitFacilitator.handleUnchanged(daughter); err != nil {
					err = errors.Wrap(err)
					return err
				}
			}

			return err
		}

		// TODO eventually remove when moving to new model of dormancy and expanded
		// tags
		if err = commitFacilitator.applyDormantAndRealizeTags(
			daughter,
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if daughter.GetGenre() == genres.Zettel {
			if err = commitFacilitator.zettelIdIndex.AddZettelId(&daughter.ObjectId); err != nil {
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

		if err = commitFacilitator.commitTransacted(daughter, mother); err != nil {
			err = errors.Wrapf(err, "Sku: %s", sku.String(daughter))
			return err
		}
	}

	if options.AddToInventoryList ||
		options.StreamIndexOptions.AddToStreamIndex {
		if store_version.GreaterOrEqual(
			commitFacilitator.storeConfig.GetConfig().GetStoreVersion(),
			store_version.V11,
		) {
			if err = markl.AssertIdIsNotNull(
				daughter.Metadata.GetObjectSig()); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}

		if err = commitFacilitator.storeConfig.AddTransacted(
			daughter,
			mother,
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = commitFacilitator.index.Add(
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

			if err = commitFacilitator.ui.TransactedNew(daughter); err != nil {
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

			if err = commitFacilitator.ui.TransactedUpdated(daughter); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}

	}

	if options.MergeCheckedOut {
		if err = commitFacilitator.ReadExternalAndMergeIfNecessary(
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

func (commitFacilitator commitFacilitator) fetchMotherIfNecessary(
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
	if err = commitFacilitator.index.ReadOneObjectId(
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
func (commitFacilitator commitFacilitator) commitTransacted(
	daughter *sku.Transacted,
	mother *sku.Transacted,
) (err error) {
	if err = commitFacilitator.inventoryListStore.AddObjectToOpenList(
		commitFacilitator.inventoryList,
		daughter,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (commitFacilitator commitFacilitator) handleUnchanged(
	object *sku.Transacted,
) (err error) {
	return commitFacilitator.ui.TransactedUnchanged(object)
}

func (commitFacilitator commitFacilitator) createType(
	typeId *ids.ObjectId,
) (err error) {
	typeObject := sku.GetTransactedPool().Get()
	defer sku.GetTransactedPool().Put(typeObject)

	switch typeId.GetGenre() {
	default:
		err = genres.MakeErrUnsupportedGenre(typeId.GetGenre())
		return err

	case genres.Type:
		typeObject.GetMetadataMutable().GetTypePtr().ResetWith(
			ids.DefaultOrPanic(genres.Type),
		)
	}

	if err = typeObject.ObjectId.SetWithIdLike(typeId); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = commitFacilitator.Commit(
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

func (commitFacilitator commitFacilitator) addTypeIfNecessary(
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

	if err = commitFacilitator.index.ObjectExists(&objectId); err == nil {
		return err
	}

	err = nil

	var typeObjectId ids.ObjectId

	if err = typeObjectId.SetWithIdLike(typeId); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = commitFacilitator.createType(&typeObjectId); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (commitFacilitator commitFacilitator) addTypeAndExpandedIfNecessary(
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
		if err = commitFacilitator.addTypeIfNecessary(tipe); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (commitFacilitator commitFacilitator) addMissingTypes(
	commitOptions sku.CommitOptions,
	object *sku.Transacted,
) (err error) {
	tipe := object.GetType()

	if !commitOptions.DontAddMissingType {
		if err = commitFacilitator.addTypeAndExpandedIfNecessary(tipe); err != nil {
			err = errors.Wrap(err)
			return err
		}
	} else {
		// TODO enforce that object has signature
	}

	return err
}
