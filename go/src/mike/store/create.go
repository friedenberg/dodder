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
func (s *Store) Reindex() (err error) {
	if !s.GetEnvRepo().GetLockSmith().IsAcquired() {
		err = file_lock.ErrLockRequired{
			Operation: "reindex",
		}

		return
	}

	if err = s.ResetIndexes(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = s.GetEnvRepo().ResetCache(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = s.GetStreamIndex().Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	missingObjects := sku.MakeList()

	for objectWithList, iterErr := range s.GetInventoryListStore().IterAllSkus() {
		if iterErr != nil {
			if env_dir.IsErrBlobMissing(iterErr) {
				missingObjects.Add(objectWithList.List)
				continue
			}

			err = errors.Wrap(iterErr)
			return
		}

		if err = s.reindexOne(objectWithList); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	s.envRepo.GetUI().Print("missing list blobs: ")

	for missingList := range missingObjects.All() {
		s.envRepo.GetUI().Print(sku.String(missingList))
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

	return store.CreateOrUpdate(external, options)
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
		err = errors.WrapExcept(err, collections.ErrExists)
		return
	}

	return
}

func (s *Store) CreateOrUpdateBlobSha(
	k interfaces.ObjectId,
	sh interfaces.Sha,
) (t *sku.Transacted, err error) {
	if !s.GetEnvRepo().GetLockSmith().IsAcquired() {
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

	if err = s.ReadOneInto(k, t); err != nil {
		if collections.IsErrNotFound(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	t.SetBlobSha(sh)

	if err = s.Commit(
		t,
		sku.CommitOptions{StoreOptions: sku.GetStoreOptionsUpdate()},
	); err != nil {
		err = errors.WrapExcept(err, collections.ErrExists)
		return
	}

	return
}

type RevertId struct {
	*ids.ObjectId
	ids.Tai
}

func (s *Store) RevertTo(
	ri RevertId,
) (err error) {
	if ri.Tai.IsEmpty() {
		return
	}

	if !s.GetEnvRepo().GetLockSmith().IsAcquired() {
		err = file_lock.ErrLockRequired{
			Operation: "update many metadata",
		}

		return
	}

	var mutter *sku.Transacted

	if mutter, err = s.GetStreamIndex().ReadOneObjectIdTai(
		ri.ObjectId,
		ri.Tai,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer sku.GetTransactedPool().Put(mutter)

	if err = s.Commit(
		mutter,
		sku.CommitOptions{StoreOptions: sku.GetStoreOptionsUpdate()},
	); err != nil {
		err = errors.WrapExcept(err, collections.ErrExists)
		return
	}

	return
}

func (s *Store) CreateOrUpdateCheckedOut(
	col sku.SkuType,
	updateCheckout bool,
) (err error) {
	external := col.GetSkuExternal()
	internal := external.GetSku()

	if !s.GetEnvRepo().GetLockSmith().IsAcquired() {
		err = file_lock.ErrLockRequired{
			Operation: fmt.Sprintf("create or update %s", internal.GetObjectId()),
		}

		return
	}

	if err = s.Commit(
		external,
		sku.CommitOptions{StoreOptions: sku.GetStoreOptionsCreate()},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if !updateCheckout {
		return
	}

	if err = s.UpdateCheckoutFromCheckedOut(
		checkout_options.OptionsWithoutMode{Force: true},
		col,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
