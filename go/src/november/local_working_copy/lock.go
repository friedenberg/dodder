package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func (u *Repo) Lock() (err error) {
	if err = u.envRepo.GetLockSmith().Lock(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO print organize files that were created if dry run or make it possible to
// commit dry-run transactions
func (u *Repo) Unlock() (err error) {
	ptl := u.PrinterTransacted()

	if u.storesInitialized {
		ui.Log().Printf("konfig has changes: %t", u.GetConfig().HasChanges())
		ui.Log().Printf("dormant has changes: %t", u.GetDormantIndex().HasChanges())

		var changes []string
		changes = append(changes, u.GetConfig().GetChanges()...)
		changes = append(changes, u.GetDormantIndex().GetChanges()...)
		u.GetStore().GetStreamIndex().SetNeedsFlushHistory(changes)

		ui.Log().Print("will flush inventory list")
		if err = u.store.FlushInventoryList(ptl); err != nil {
			err = errors.Wrap(err)
			return
		}

		ui.Log().Print("will flush store")
		if err = u.store.Flush(
			u.PrinterHeader(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		ui.Log().Print("will flush konfig")
		if err = u.config.Flush(
			u.GetEnvRepo(),
			u.GetStore().GetTypedBlobStore(),
			u.PrinterHeader(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		ui.Log().Print("will flush dormant")
		if err = u.dormantIndex.Flush(
			u.GetEnvRepo(),
			u.PrinterHeader(),
			u.config.GetCLIConfig().IsDryRun(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		u.GetStore().GetStreamIndex().SetNeedsFlushHistory(changes)

		wg := errors.MakeWaitGroupParallel()
		wg.Do(
			func() error {
				ui.Log().Print("will flush store second time")
				// second store flush is necessary because of konfig changes
				return u.store.Flush(
					u.PrinterHeader(),
				)
			},
		)

		if err = wg.GetError(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	// explicitly do not unlock if there was an error to encourage user
	// interaction
	// and manual recovery
	if err = u.envRepo.GetLockSmith().Unlock(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
