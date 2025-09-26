package store

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/file_lock"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

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
		return err
	}

	return err
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
		return err
	}

	return err
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

		return object, err
	}

	object = sku.GetTransactedPool().Get()

	if err = object.ObjectId.SetWithIdLike(objectId); err != nil {
		err = errors.Wrap(err)
		return object, err
	}

	if err = store.ReadOneInto(objectId, object); err != nil {
		if collections.IsErrNotFound(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return object, err
		}
	}

	object.SetBlobDigest(blobDigest)

	if err = store.Commit(
		object,
		sku.CommitOptions{StoreOptions: sku.GetStoreOptionsUpdate()},
	); err != nil {
		err = errors.WrapExceptSentinel(err, collections.ErrExists)
		return object, err
	}

	return object, err
}

type RevertId struct {
	*ids.ObjectId
	ids.Tai
}

func (store *Store) RevertTo(
	revertId RevertId,
) (err error) {
	if revertId.Tai.IsEmpty() {
		return err
	}

	if !store.GetEnvRepo().GetLockSmith().IsAcquired() {
		err = file_lock.ErrLockRequired{
			Operation: "update many metadata",
		}

		return err
	}

	var mother *sku.Transacted

	if mother, err = store.streamIndex.ReadOneObjectIdTai(
		revertId.ObjectId,
		revertId.Tai,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer sku.GetTransactedPool().Put(mother)

	if err = store.Commit(
		mother,
		sku.CommitOptions{StoreOptions: sku.GetStoreOptionsUpdate()},
	); err != nil {
		err = errors.WrapExceptSentinel(err, collections.ErrExists)
		return err
	}

	return err
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

		return err
	}

	if err = store.Commit(
		external,
		sku.CommitOptions{StoreOptions: sku.GetStoreOptionsCreate()},
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if !updateCheckout {
		return err
	}

	if err = store.UpdateCheckoutFromCheckedOut(
		checkout_options.OptionsWithoutMode{Force: true},
		col,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
