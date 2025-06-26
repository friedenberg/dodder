package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/lima/store_fs"
)

func (u *Repo) DeleteFiles(fs interfaces.Collection[*fd.FD]) (err error) {
	deleteOp := store_fs.DeleteCheckout{}

	if err = deleteOp.Run(
		u.GetConfig().GetCLIConfig().IsDryRun(),
		u.GetEnvRepo(),
		u.PrinterFDDeleted(),
		fs,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
