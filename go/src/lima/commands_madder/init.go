package commands_madder

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/india/command_components_madder"
)

func init() {
	utility.AddCmd("init", &Init{
		tipe: ids.GetOrPanic(ids.TypeTomlBlobStoreConfigVCurrent).Type,
		blobStoreConfig: &blob_store_configs.DefaultType{
			CompressionType:   compression_type.CompressionTypeDefault,
			LockInternalFiles: true,
		},
	})

	utility.AddCmd("init-pointer", &Init{
		tipe: ids.GetOrPanic(
			ids.TypeTomlBlobStoreConfigPointerV0,
		).Type,
		blobStoreConfig: &blob_store_configs.TomlPointerV0{},
	})

	utility.AddCmd("init-sftp-explicit", &Init{
		tipe: ids.GetOrPanic(
			ids.TypeTomlBlobStoreConfigSftpExplicitV0,
		).Type,
		blobStoreConfig: &blob_store_configs.TomlSFTPV0{},
	})

	utility.AddCmd("init-sftp-ssh_config", &Init{
		tipe: ids.GetOrPanic(
			ids.TypeTomlBlobStoreConfigSftpViaSSHConfigV0,
		).Type,
		blobStoreConfig: &blob_store_configs.TomlSFTPViaSSHConfigV0{},
	})
}

type Init struct {
	tipe            ids.Type
	blobStoreConfig blob_store_configs.ConfigMutable

	command_components_madder.EnvBlobStore
	command_components_madder.Init
}

var _ interfaces.CommandComponentWriter = (*Init)(nil)

func (cmd *Init) SetFlagDefinitions(
	flagDefinitions interfaces.CLIFlagDefinitions,
) {
	cmd.blobStoreConfig.SetFlagDefinitions(flagDefinitions)
}

func (cmd *Init) Run(req command.Request) {
	var blobStoreName ids.Tag

	if err := blobStoreName.Set(req.PopArg("blob store name")); err != nil {
		errors.ContextCancelWithBadRequestError(req, err)
	}

	req.AssertNoMoreArgs()

	envBlobStore := cmd.MakeEnvBlobStore(req)

	pathConfig := cmd.InitBlobStore(
		req,
		envBlobStore,
		blobStoreName.String(),
		&blob_store_configs.TypedConfig{
			Type: cmd.tipe,
			Blob: cmd.blobStoreConfig,
		},
	)

	envBlobStore.GetUI().Printf("Wrote config to %s", pathConfig)
}
