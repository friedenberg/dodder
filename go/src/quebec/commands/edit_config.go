package commands

import (
	"fmt"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/vim_cli_options_builder"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
	"code.linenisgreat.com/dodder/go/src/papa/user_ops"
)

func init() {
	command.Register("edit-config", &EditConfig{})
}

type EditConfig struct {
	command_components.LocalWorkingCopy
}

func (cmd EditConfig) Run(
	req command.Request,
) {
	args := req.PopArgs()
	localWorkingCopy := cmd.MakeLocalWorkingCopy(req)

	if len(args) > 0 {
		ui.Err().Print("Command edit-konfig ignores passed in arguments.")
	}

	var sk *sku.Transacted

	{
		var err error

		if sk, err = cmd.editInVim(localWorkingCopy); err != nil {
			localWorkingCopy.Cancel(err)
		}
	}

	localWorkingCopy.Must(
		errors.MakeFuncContextFromFuncErr(localWorkingCopy.Reset),
	)

	localWorkingCopy.Must(
		errors.MakeFuncContextFromFuncErr(localWorkingCopy.Lock),
	)

	if err := localWorkingCopy.GetStore().CreateOrUpdateDefaultProto(
		sk,
		sku.StoreOptions{},
	); err != nil {
		localWorkingCopy.Cancel(err)
	}

	localWorkingCopy.Must(
		errors.MakeFuncContextFromFuncErr(localWorkingCopy.Unlock),
	)
}

func (cmd EditConfig) editInVim(
	repo *local_working_copy.Repo,
) (sk *sku.Transacted, err error) {
	var file *os.File

	if file, err = repo.GetEnvRepo().GetTempLocal().FileTempWithTemplate(
		fmt.Sprintf("*.%s", repo.GetConfig().GetFileExtensions().GetFileExtensionConfig()),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	path := file.Name()

	if err = file.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = cmd.makeTempConfigFile(repo, path); err != nil {
		err = errors.Wrap(err)
		return
	}

	openVimOp := user_ops.OpenEditor{
		VimOptions: vim_cli_options_builder.New().
			WithFileType("dodder-object").
			Build(),
	}

	if err = openVimOp.Run(repo, path); err != nil {
		err = errors.Wrap(err)
		return
	}

	if sk, err = cmd.readTempConfigFile(repo, path); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (cmd EditConfig) makeTempConfigFile(
	repo *local_working_copy.Repo,
	path string,
) (err error) {
	var k *sku.Transacted

	if k, err = repo.GetStore().ReadTransactedFromObjectId(&ids.Config{}); err != nil {
		err = errors.Wrap(err)
		return
	}

	var i sku.FSItem
	i.Reset()

	if err = i.Object.Set(path); err != nil {
		err = errors.Wrap(err)
		return
	}

	i.FDs.Add(&i.Object)

	if err = repo.GetEnvWorkspace().GetStoreFS().GetFileEncoder().Encode(
		checkout_options.TextFormatterOptions{},
		k,
		&i,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (cmd EditConfig) readTempConfigFile(
	localWorkingCopy *local_working_copy.Repo,
	path string,
) (sk *sku.Transacted, err error) {
	sk = sku.GetTransactedPool().Get()

	if sk.ObjectId.Set("konfig"); err != nil {
		err = errors.Wrap(err)
		return
	}

	var file *os.File

	if file, err = files.Open(path); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, file)

	if err = localWorkingCopy.GetEnvWorkspace().GetStoreFS().ReadOneExternalObjectReader(
		file,
		sk,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
