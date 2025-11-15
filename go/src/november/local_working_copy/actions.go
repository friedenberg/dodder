package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/lima/store_fs"
)

func (local *Repo) DeleteFiles(fs interfaces.Collection[*fd.FD]) (err error) {
	deleteOp := store_fs.DeleteCheckout{}

	if err = deleteOp.Run(
		local.GetConfig().IsDryRun(),
		local.GetEnvRepo(),
		local.PrinterFDDeleted(),
		fs,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
