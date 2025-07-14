package commands

import (
	"flag"
	"fmt"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("blob_store-init", &BlobStoreInit{
		tipe: ids.GetOrPanic(ids.TypeTomlBlobStoreConfigV0).Type,
		blobStoreConfig: &blob_store_configs.TomlV0{
			CompressionType:   compression_type.CompressionTypeDefault,
			LockInternalFiles: true,
		},
	})

	command.Register("blob_store-init-sftp", &BlobStoreInit{
		tipe:            ids.GetOrPanic(ids.TypeTomlBlobStoreConfigV0).Type,
		blobStoreConfig: &blob_store_configs.TomlSftpV0{},
	})
}

type BlobStoreInit struct {
	tipe            ids.Type
	blobStoreConfig blob_store_configs.ConfigMutable

	command_components.EnvRepo
}

func (cmd *BlobStoreInit) SetFlagSet(flagSet *flag.FlagSet) {
	cmd.blobStoreConfig.SetFlagSet(flagSet)
	cmd.EnvRepo.SetFlagSet(flagSet)
}

func (cmd *BlobStoreInit) Run(req command.Request) {
	// TODO validate no space
	blobStoreName := req.PopArg("blob store name")

	req.AssertNoMoreArgs()

	env := cmd.MakeEnvRepo(req, false)

	blobStoreCount := len(env.GetBlobStores())

	dir := env.DirBlobStoreConfigs()

	if err := env.MakeDir(dir); err != nil {
		env.CancelWithError(err)
		return
	}

	if err := triple_hyphen_io.EncodeToFile(
		blob_store_configs.Coder,
		&triple_hyphen_io.TypedBlob[blob_store_configs.Config]{
			Type: cmd.tipe,
			Blob: cmd.blobStoreConfig,
		},
		filepath.Join(
			dir,
			fmt.Sprintf("%d-%s.%s",
				blobStoreCount,
				blobStoreName,
				env_repo.FileNameBlobStoreConfig,
			),
		),
	); err != nil {
		env.CancelWithError(err)
		return
	}
}
