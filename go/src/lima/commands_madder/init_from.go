package commands_madder

import (
	"path/filepath"

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

	var typedConfig triple_hyphen_io.TypedBlob[blob_store_configs.Config]

	{
		var err error

		if typedConfig, err = triple_hyphen_io.DecodeFromFile(
			blob_store_configs.Coder,
			configPathFD.String(),
		); err != nil {
			envBlobStore.Cancel(err)
			return
		}
	}

	for {
		configUpgraded, ok := typedConfig.Blob.(blob_store_configs.ConfigUpgradeable)

		if !ok {
			break
		}

		typedConfig.Blob, typedConfig.Type = configUpgraded.Upgrade()
	}

	cmd.tryAddBasePath(req, typedConfig.Blob)

	pathConfig := cmd.InitBlobStore(
		req,
		envBlobStore,
		name.String(),
		&typedConfig,
	)

	envBlobStore.GetUI().Printf("Wrote config to %s", pathConfig)
}

func (cmd InitFrom) tryAddBasePath(
	req command.Request,
	config blob_store_configs.Config,
) {
	var basePath string

	{
		var ok bool

		basePath, ok = blob_store_configs.GetBasePath(config)

		if !ok {
			return
		}
	}

	if filepath.IsAbs(basePath) {
		return
	}

	var configLocalMutable blob_store_configs.ConfigLocalMutable

	{
		var ok bool

		configLocalMutable, ok = config.(blob_store_configs.ConfigLocalMutable)

		if !ok {
			errors.ContextCancelWithBadRequestf(
				req,
				"expected %T but got %T",
				configLocalMutable,
				config,
			)

			return
		}

		// TODO get base path relative to config path

		// TODO populate basepath for config to be absolute
		configLocalMutable.SetBasePath("test")
	}
}
