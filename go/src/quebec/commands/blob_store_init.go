package commands

import (
	"fmt"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

// TODO move to madder
func init() {
	command.Register("blob_store-init", &BlobStoreInit{
		tipe: ids.GetOrPanic(ids.TypeTomlBlobStoreConfigV1).Type,
		blobStoreConfig: &blob_store_configs.TomlV1{
			CompressionType:   compression_type.CompressionTypeDefault,
			LockInternalFiles: true,
		},
	})

	command.Register("blob_store-init-sftp-explicit", &BlobStoreInit{
		tipe: ids.GetOrPanic(
			ids.TypeTomlBlobStoreConfigSftpExplicitV0,
		).Type,
		blobStoreConfig: &blob_store_configs.TomlSFTPV0{},
	})

	command.Register("blob_store-init-sftp-ssh_config", &BlobStoreInit{
		tipe: ids.GetOrPanic(
			ids.TypeTomlBlobStoreConfigSftpViaSSHConfigV0,
		).Type,
		blobStoreConfig: &blob_store_configs.TomlSFTPViaSSHConfigV0{},
	})
}

type BlobStoreInit struct {
	tipe            ids.Type
	blobStoreConfig blob_store_configs.ConfigMutable

	command_components.EnvRepo
}

func (cmd *BlobStoreInit) SetFlagSet(flagSet interfaces.CommandLineFlagDefinitions) {
	cmd.blobStoreConfig.SetFlagSet(flagSet)
}

func (cmd *BlobStoreInit) Run(req command.Request) {
	// TODO validate no space
	var blobStoreName ids.Tag

	if err := blobStoreName.Set(req.PopArg("blob store name")); err != nil {
		errors.ContextCancelWithBadRequestError(req, err)
	}

	req.AssertNoMoreArgs()

	env := cmd.MakeEnvRepo(req, false)

	blobStoreCount := len(env.GetBlobStores())

	dir := env.DirBlobStoreConfigs()

	if err := env.MakeDir(dir); err != nil {
		env.Cancel(err)
		return
	}

	pathConfig := filepath.Join(
		dir,
		fmt.Sprintf("%d-%s.%s",
			blobStoreCount,
			blobStoreName,
			env_repo.FileNameBlobStoreConfig,
		),
	)

	if err := triple_hyphen_io.EncodeToFile(
		blob_store_configs.Coder,
		&triple_hyphen_io.TypedBlob[blob_store_configs.Config]{
			Type: cmd.tipe,
			Blob: cmd.blobStoreConfig,
		},
		pathConfig,
	); err != nil {
		env.Cancel(err)
		return
	}

	env.GetUI().Printf("Wrote config to %s", pathConfig)
}
