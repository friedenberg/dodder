package commands

import (
	"flag"
	"path/filepath"
	"strconv"

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
	req.AssertNoMoreArgs()

	env := cmd.MakeEnvRepo(req, false)

	blobStoreCount := len(env.GetBlobStores())

	dir := env.DirBlobStores(
		strconv.Itoa(blobStoreCount),
	)

	if err := env.MakeDir(dir); err != nil {
		env.CancelWithError(err)
		return
	}

	triple_hyphen_io.EncodeToFile(
		env,
		blob_store_configs.Coder,
		&triple_hyphen_io.TypedBlob[blob_store_configs.Config]{
			Type: cmd.tipe,
			Blob: cmd.blobStoreConfig,
		},
		filepath.Join(
			dir,
			env_repo.FileNameBlobStoreConfig,
		),
	)
}
