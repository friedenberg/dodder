package commands

import (
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/vim_cli_options_builder"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/repo_config_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
	"code.linenisgreat.com/dodder/go/src/papa/user_ops"
)

func init() {
	command.Register("dormant-edit", &DormantEdit{})
}

type DormantEdit struct {
	command_components.LocalWorkingCopy
}

func (cmd DormantEdit) Run(dep command.Request) {
	args := dep.PopArgs()
	localWorkingCopy := cmd.MakeLocalWorkingCopy(dep)

	if len(args) > 0 {
		ui.Err().Print("Command dormant-edit ignores passed in arguments.")
	}

	var sh interfaces.Sha

	{
		var err error

		if sh, err = cmd.editInVim(localWorkingCopy); err != nil {
			localWorkingCopy.CancelWithError(err)
			return
		}
	}

	if err := localWorkingCopy.Reset(); err != nil {
		localWorkingCopy.CancelWithError(err)
		return
	}

	if err := localWorkingCopy.Lock(); err != nil {
		localWorkingCopy.CancelWithError(err)
		return
	}

	defer localWorkingCopy.Must(localWorkingCopy.Unlock)

	if _, err := localWorkingCopy.GetStore().UpdateKonfig(sh); err != nil {
		localWorkingCopy.CancelWithError(err)
		return
	}
}

// TODO refactor into common
func (c DormantEdit) editInVim(
	u *local_working_copy.Repo,
) (sh interfaces.Sha, err error) {
	var p string

	if p, err = c.makeTempKonfigFile(u); err != nil {
		err = errors.Wrap(err)
		return
	}

	openVimOp := user_ops.OpenEditor{
		VimOptions: vim_cli_options_builder.New().
			Build(),
	}

	if err = openVimOp.Run(u, p); err != nil {
		err = errors.Wrap(err)
		return
	}

	if sh, err = c.readTempKonfigFile(u, p); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO refactor into common
func (c DormantEdit) makeTempKonfigFile(
	u *local_working_copy.Repo,
) (p string, err error) {
	var k *sku.Transacted

	if k, err = u.GetStore().ReadTransactedFromObjectId(&ids.Config{}); err != nil {
		err = errors.Wrap(err)
		return
	}

	var f *os.File

	if f, err = u.GetEnvRepo().GetTempLocal().FileTemp(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, f)

	p = f.Name()

	format := u.GetStore().GetConfigBlobFormat()

	if _, err = format.FormatSavedBlob(f, k.GetBlobSha()); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO refactor into common
func (c DormantEdit) readTempKonfigFile(
	u *local_working_copy.Repo,
	p string,
) (sh interfaces.Sha, err error) {
	var f *os.File

	if f, err = files.Open(p); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, f)

	format := u.GetStore().GetConfigBlobFormat()

	var k repo_config_blobs.V0

	var aw interfaces.ShaWriteCloser

	if aw, err = u.GetEnvRepo().BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, aw)

	// TODO-P3 offer option to edit again
	if _, err = format.DecodeFrom(&k, io.TeeReader(f, aw)); err != nil {
		err = errors.Wrap(err)
		return
	}

	sh = aw.GetShaLike()

	return
}
