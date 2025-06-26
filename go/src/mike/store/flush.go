package store

import (
	"encoding/gob"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/inventory_list_store"
)

func (store *Store) FlushInventoryList(
	p interfaces.FuncIter[*sku.Transacted],
) (err error) {
	if store.GetConfig().GetCLIConfig().IsDryRun() {
		return
	}

	if !store.GetEnvRepo().GetLockSmith().IsAcquired() {
		return
	}

	ui.Log().Printf("saving inventory list")

	var inventoryListSku *sku.Transacted

	store.inventoryList.Description = store.GetConfig().GetCLIConfig().Description

	if inventoryListSku, err = store.GetInventoryListStore().Create(
		store.inventoryList,
	); err != nil {
		if errors.Is(err, inventory_list_store.ErrEmpty) {
			ui.Log().Printf("inventory list was empty")
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	if inventoryListSku != nil {
		if err = store.GetStreamIndex().Add(
			inventoryListSku,
			sku.CommitOptions{
				StoreOptions: sku.StoreOptions{},
			},
		); err != nil {
			err = errors.Wrap(err)
			return
		}
		defer sku.GetTransactedPool().Put(inventoryListSku)

		if store.GetConfig().GetCLIConfig().PrintOptions.PrintInventoryLists {
			if err = p(inventoryListSku); err != nil {
				err = errors.Wrap(err)
				return
			}
		}
	}

	if store.inventoryList, err = store.inventoryListStore.MakeOpenList(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.GetInventoryListStore().Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	ui.Log().Printf("done saving inventory list")

	return
}

func (store *Store) Flush(
	printerHeader interfaces.FuncIter[string],
) (err error) {
	// TODO handle flushes with dry run
	if store.GetConfig().GetCLIConfig().IsDryRun() {
		return
	}

	wg := errors.MakeWaitGroupParallel()

	if store.GetEnvRepo().GetLockSmith().IsAcquired() {
		gob.Register(quiter.StringerKeyerPtr[ids.Type, *ids.Type]{}) // TODO check if can be removed
		wg.Do(func() error { return store.streamIndex.Flush(printerHeader) })
		wg.Do(store.GetAbbrStore().Flush)
		wg.Do(store.zettelIdIndex.Flush)
		wg.Do(store.Abbr.Flush)
	}

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
