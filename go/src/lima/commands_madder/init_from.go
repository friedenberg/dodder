package commands_madder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/india/command_components_madder"
)

func init() {
	utility.AddCmd("init-from", &InitFrom{})
}

type InitFrom struct {
	command_components_madder.EnvBlobStore
	command_components_madder.Init
}

var _ interfaces.CommandComponentWriter = (*InitFrom)(nil)

func (cmd *InitFrom) SetFlagDefinitions(
	flagDefinitions interfaces.CLIFlagDefinitions,
) {
}

func (cmd InitFrom) Complete(
	req command.Request,
	envLocal env_local.Env,
	commandLine command.CommandLine,
) {
	// TODO support completion for config path
}

func (cmd *InitFrom) Run(req command.Request) {
	var name ids.Tag

	if err := name.Set(req.PopArg("blob store name")); err != nil {
		errors.ContextCancelWithBadRequestError(req, err)
	}

	var configPathFD fd.FD

	if err := configPathFD.Set(req.PopArg("blob store config path")); err != nil {
		errors.ContextCancelWithBadRequestError(req, err)
	}

	req.AssertNoMoreArgs()

	envBlobStore := cmd.MakeEnvBlobStore(req)

	var config triple_hyphen_io.TypedBlob[blob_store_configs.Config]

	{
		var err error

		if config, err = triple_hyphen_io.DecodeFromFile(
			blob_store_configs.Coder,
			configPathFD.String(),
		); err != nil {
			envBlobStore.Cancel(err)
			return
		}
	}

	if config, ok := config.Blob.(blob_store_configs.ConfigLocal); ok {
		config, ok := config.(blob_store_configs.ConfigLocalMutable)

		if !ok {
			// TODO emit error, blob must support setting base path
		}

		// TODO get base path relative to config path
		config.SetBasePath("test")
	}
	// TODO populate basepath for config to be absolute

	pathConfig := cmd.InitBlobStore(
		req,
		envBlobStore,
		name.String(),
		&config,
	)

	envBlobStore.GetUI().Printf("Wrote config to %s", pathConfig)
}
