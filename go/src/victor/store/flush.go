package store

import (
	"encoding/gob"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
	"code.linenisgreat.com/dodder/go/src/oscar/inventory_list_store"
)

func (store *Store) FlushInventoryList(
	p interfaces.FuncIter[*sku.Transacted],
) (err error) {
	if store.GetConfigStore().GetConfig().IsDryRun() {
		return err
	}

	if !store.GetEnvRepo().GetLockSmith().IsAcquired() {
		return err
	}

	ui.Log().Printf("saving inventory list")

	var inventoryListSku *sku.Transacted

	store.inventoryList.GetDescriptionMutable().ResetWith(
		store.GetConfigStoreMutable().GetConfig().Description,
	)

	if inventoryListSku, err = store.GetInventoryListStore().Create(
		store.inventoryList,
	); err != nil {
		if errors.Is(err, inventory_list_store.ErrEmptyInventoryList) {
			ui.Log().Printf("inventory list was empty")
			err = nil
		} else {
			err = errors.Wrap(err)
			return err
		}
	} else {
		defer sku.GetTransactedPool().Put(inventoryListSku)
	}

	if inventoryListSku != nil {
		if err = store.streamIndex.Add(
			inventoryListSku,
			sku.CommitOptions{
				StoreOptions: sku.StoreOptions{},
			},
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if store.GetConfigStore().GetConfig().GetPrintOptions().PrintInventoryLists {
			if err = p(inventoryListSku); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}
	}

	if store.inventoryList, err = store.inventoryListStore.MakeOpenList(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = store.GetInventoryListStore().Flush(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	ui.Log().Printf("done saving inventory list")

	return err
}

func (store *Store) Flush(
	printerHeader interfaces.FuncIter[string],
) (err error) {
	// TODO handle flushes with dry run
	if store.GetConfigStore().GetConfig().IsDryRun() {
		return err
	}

	wg := errors.MakeWaitGroupParallel()

	if store.GetEnvRepo().GetLockSmith().IsAcquired() {
		gob.Register(
			quiter.StringerKeyerPtr[ids.Type, *ids.Type]{},
		) // TODO check if can be removed
		wg.Do(func() error { return store.streamIndex.Flush(printerHeader) })
		wg.Do(store.GetAbbrStore().Flush)
		wg.Do(store.zettelIdIndex.Flush)
		wg.Do(store.Abbr.Flush)
	}

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
