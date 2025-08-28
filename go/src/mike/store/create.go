package store

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/file_lock"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
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

	missingObjects := sku.MakeListTransacted()

	for objectWithList, iterErr := range store.GetInventoryListStore().IterAllSkus() {
		if iterErr != nil {
			if env_dir.IsErrBlobMissing(iterErr) {
				missingObjects.Add(objectWithList.List)
				continue
			}

			err = errors.Wrapf(
				iterErr,
				"List: %s",
				sku.String(objectWithList.List),
			)
			return
		}

		if err = store.reindexOne(objectWithList); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	store.envRepo.GetUI().Print("missing list blobs: ")

	for missingList := range missingObjects.All() {
		store.envRepo.GetUI().Print(sku.String(missingList))
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

func (store *Store) CreateOrUpdateBlobSha(
	k interfaces.ObjectId,
	sh interfaces.BlobId,
) (t *sku.Transacted, err error) {
	if !store.GetEnvRepo().GetLockSmith().IsAcquired() {
		err = file_lock.ErrLockRequired{
			Operation: fmt.Sprintf(
				"create or update %s",
				k.GetGenre(),
			),
		}

		return
	}

	t = sku.GetTransactedPool().Get()

	if err = t.ObjectId.SetWithIdLike(k); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.ReadOneInto(k, t); err != nil {
		if collections.IsErrNotFound(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	t.SetBlobDigest(sh)

	if err = store.Commit(
		t,
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
	ri RevertId,
) (err error) {
	if ri.Tai.IsEmpty() {
		return
	}

	if !store.GetEnvRepo().GetLockSmith().IsAcquired() {
		err = file_lock.ErrLockRequired{
			Operation: "update many metadata",
		}

		return
	}

	var mutter *sku.Transacted

	if mutter, err = store.GetStreamIndex().ReadOneObjectIdTai(
		ri.ObjectId,
		ri.Tai,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer sku.GetTransactedPool().Put(mutter)

	if err = store.Commit(
		mutter,
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
