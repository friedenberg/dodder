package store

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/file_lock"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

// TODO-P2 add support for quiet reindexing
func (store *Store) Reindex() (err error) {
	if !store.GetEnvRepo().GetLockSmith().IsAcquired() {
		err = file_lock.ErrLockRequired{
			Operation: "reindex",
		}

		return
	}

	if err = store.ResetIndexes(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.GetEnvRepo().ResetCache(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.GetStreamIndex().Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	type objectWithError struct {
		error
		sku.ObjectWithList
	}

	objectsWithErrors := make(map[string]objectWithError)
	unidentifiedErrors := make([]error, 0)

	for objectWithList, iterErr := range store.GetInventoryListStore().AllInventoryListObjectsAndContents() {
		if iterErr != nil {
			if objectWithList.List == nil {
				unidentifiedErrors = append(unidentifiedErrors, iterErr)
			} else {
				keyBytes := objectWithList.List.GetObjectDigest().GetBytes()

				objectsWithErrors[string(keyBytes)] = objectWithError{
					error: iterErr,
					ObjectWithList: sku.ObjectWithList{
						List: objectWithList.List.CloneTransacted(),
					},
				}
			}

			continue
		}

		if err = store.reindexOne(objectWithList); err != nil {
			keyBytes := objectWithList.List.GetObjectDigest().GetBytes()

			objectsWithErrors[string(keyBytes)] = objectWithError{
				error: err,
				ObjectWithList: sku.ObjectWithList{
					Object: objectWithList.Object.CloneTransacted(),
					List:   objectWithList.List.CloneTransacted(),
				},
			}

			continue
		}
	}

	store.envRepo.GetUI().Print("unidentified errors:")

	for _, err := range unidentifiedErrors {
		ui.CLIErrorTreeEncoder.EncodeTo(err, store.envRepo.GetUI())
	}

	store.envRepo.GetUI().Print("objects with errors:")

	for _, objectWithError := range objectsWithErrors {
		ui.CLIErrorTreeEncoder.EncodeTo(err, store.envRepo.GetUI())

		if objectWithError.Object == nil {
			store.envRepo.GetUI().Printf(
				"Error: %s, List: %q",
				objectWithError.error,
				sku.String(objectWithError.List),
			)
		} else {
			store.envRepo.GetUI().Printf(
				"Error: %s, List: %q, Object: %q",
				objectWithError.error,
				sku.String(objectWithError.List),
				sku.String(objectWithError.Object),
			)
		}
	}

	return
}

func (store *Store) CreateOrUpdateDefaultProto(
	external sku.ExternalLike,
	storeOptions sku.StoreOptions,
) (err error) {
	options := sku.CommitOptions{
		Proto:        store.protoZettel,
		StoreOptions: storeOptions,
	}

	if err = store.CreateOrUpdate(external, options); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) CreateOrUpdate(
	external sku.ExternalLike,
	options sku.CommitOptions,
) (err error) {
	options.AddToInventoryList = true
	options.UpdateTai = true
	options.RunHooks = true
	options.Validate = true

	if err = store.Commit(
		external,
		options,
	); err != nil {
		err = errors.WrapExceptSentinel(err, collections.ErrExists)
		return
	}

	return
}

func (store *Store) CreateOrUpdateBlobDigest(
	objectId interfaces.ObjectId,
	blobDigest interfaces.MarklId,
) (object *sku.Transacted, err error) {
	if !store.GetEnvRepo().GetLockSmith().IsAcquired() {
		err = file_lock.ErrLockRequired{
			Operation: fmt.Sprintf(
				"create or update %s",
				objectId.GetGenre(),
			),
		}

		return
	}

	object = sku.GetTransactedPool().Get()

	if err = object.ObjectId.SetWithIdLike(objectId); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.ReadOneInto(objectId, object); err != nil {
		if collections.IsErrNotFound(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	object.SetBlobDigest(blobDigest)

	if err = store.Commit(
		object,
		sku.CommitOptions{StoreOptions: sku.GetStoreOptionsUpdate()},
	); err != nil {
		err = errors.WrapExceptSentinel(err, collections.ErrExists)
		return
	}

	return
}

type RevertId struct {
	*ids.ObjectId
	ids.Tai
}

func (store *Store) RevertTo(
	revertId RevertId,
) (err error) {
	if revertId.Tai.IsEmpty() {
		return
	}

	if !store.GetEnvRepo().GetLockSmith().IsAcquired() {
		err = file_lock.ErrLockRequired{
			Operation: "update many metadata",
		}

		return
	}

	var mother *sku.Transacted

	if mother, err = store.GetStreamIndex().ReadOneObjectIdTai(
		revertId.ObjectId,
		revertId.Tai,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer sku.GetTransactedPool().Put(mother)

	if err = store.Commit(
		mother,
		sku.CommitOptions{StoreOptions: sku.GetStoreOptionsUpdate()},
	); err != nil {
		err = errors.WrapExceptSentinel(err, collections.ErrExists)
		return
	}

	return
}

func (store *Store) CreateOrUpdateCheckedOut(
	col sku.SkuType,
	updateCheckout bool,
) (err error) {
	external := col.GetSkuExternal()
	internal := external.GetSku()

	if !store.GetEnvRepo().GetLockSmith().IsAcquired() {
		err = file_lock.ErrLockRequired{
			Operation: fmt.Sprintf(
				"create or update %s",
				internal.GetObjectId(),
			),
		}

		return
	}

	if err = store.Commit(
		external,
		sku.CommitOptions{StoreOptions: sku.GetStoreOptionsCreate()},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if !updateCheckout {
		return
	}

	if err = store.UpdateCheckoutFromCheckedOut(
		checkout_options.OptionsWithoutMode{Force: true},
		col,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
