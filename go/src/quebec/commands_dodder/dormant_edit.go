package commands_dodder

import (
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/vim_cli_options_builder"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components_dodder"
	"code.linenisgreat.com/dodder/go/src/papa/user_ops"
)

func init() {
	utility.AddCmd("dormant-edit", &DormantEdit{})
}

type DormantEdit struct {
	command_components_dodder.LocalWorkingCopy
}

func (cmd DormantEdit) Run(req command.Request) {
	args := req.PopArgs()
	localWorkingCopy := cmd.MakeLocalWorkingCopy(req)

	if len(args) > 0 {
		ui.Err().Print("Command dormant-edit ignores passed in arguments.")
	}

	var sh interfaces.MarklId

	{
		var err error

		if sh, err = cmd.editInVim(localWorkingCopy); err != nil {
			localWorkingCopy.Cancel(err)
			return
		}
	}

	if err := localWorkingCopy.Reset(); err != nil {
		localWorkingCopy.Cancel(err)
		return
	}

	if err := localWorkingCopy.Lock(); err != nil {
		localWorkingCopy.Cancel(err)
		return
	}

	defer localWorkingCopy.Must(
		errors.MakeFuncContextFromFuncErr(localWorkingCopy.Unlock),
	)

	if _, err := localWorkingCopy.GetStore().UpdateKonfig(sh); err != nil {
		localWorkingCopy.Cancel(err)
		return
	}
}

// TODO refactor into common
func (cmd DormantEdit) editInVim(
	u *local_working_copy.Repo,
) (sh interfaces.MarklId, err error) {
	var p string

	if p, err = cmd.makeTempKonfigFile(u); err != nil {
		err = errors.Wrap(err)
		return sh, err
	}

	openVimOp := user_ops.OpenEditor{
		VimOptions: vim_cli_options_builder.New().
			Build(),
	}

	if err = openVimOp.Run(u, p); err != nil {
		err = errors.Wrap(err)
		return sh, err
	}

	if sh, err = cmd.readTempKonfigFile(u, p); err != nil {
		err = errors.Wrap(err)
		return sh, err
	}

	return sh, err
}

// TODO refactor into common
func (cmd DormantEdit) makeTempKonfigFile(
	repo *local_working_copy.Repo,
) (path string, err error) {
	var object *sku.Transacted

	if object, err = repo.GetStore().ReadTransactedFromObjectId(&ids.Config{}); err != nil {
		err = errors.Wrap(err)
		return path, err
	}

	var file *os.File

	if file, err = repo.GetEnvRepo().GetTempLocal().FileTemp(); err != nil {
		err = errors.Wrap(err)
		return path, err
	}

	defer errors.DeferredCloser(&err, file)

	var readCloser io.ReadCloser

	if readCloser, err = repo.GetEnvRepo().GetDefaultBlobStore().MakeBlobReader(
		object.GetBlobDigest(),
	); err != nil {
		err = errors.Wrap(err)
		return path, err
	}

	path = file.Name()

	if _, err = ohio.CopyBuffered(file, readCloser); err != nil {
		err = errors.Wrap(err)
		return path, err
	}

	return path, err
}

// TODO refactor into common
func (cmd DormantEdit) readTempKonfigFile(
	repo *local_working_copy.Repo,
	path string,
) (sh interfaces.MarklId, err error) {
	var file *os.File

	if file, err = files.Open(path); err != nil {
		err = errors.Wrap(err)
		return sh, err
	}

	defer errors.DeferredCloser(&err, file)

	var writeCloser interfaces.BlobWriter

	if writeCloser, err = repo.GetEnvRepo().GetDefaultBlobStore().MakeBlobWriter(nil); err != nil {
		err = errors.Wrap(err)
		return sh, err
	}

	defer errors.DeferredCloser(&err, writeCloser)

	var typedBlob repo_configs.TypedBlob

	coder := repo.GetStore().GetConfigBlobCoder()

	// TODO-P3 offer option to edit again
	if _, err = coder.DecodeFrom(
		&typedBlob,
		io.TeeReader(file, writeCloser),
	); err != nil {
		err = errors.Wrap(err)
		return sh, err
	}

	// TODO persist blob type

	sh = writeCloser.GetMarklId()

	return sh, err
}
