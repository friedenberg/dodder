package command_components

import (
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
)

// TODO migrate to using `command_components.BlobStore`
type RemoteBlobStore struct {
	BasePath string
	Config   blob_store_configs.TomlV0
}

func (cmd *RemoteBlobStore) SetFlagSet(flagSet *flags.FlagSet) {
	cmd.Config.CompressionType = compression_type.CompressionTypeDefault
	cmd.Config.CompressionType.SetFlagSet(flagSet)
	flagSet.StringVar(&cmd.BasePath, "blobs", "", "")
}

func (cmd *RemoteBlobStore) GetBlobStoreConfigNamed() blob_stores.BlobStoreConfigNamed {
	return blob_stores.BlobStoreConfigNamed{
		BasePath: cmd.BasePath,
		Config:   &cmd.Config,
	}
}

func (cmd *RemoteBlobStore) MakeRemoteBlobStore(
	envLocal env_local.Env,
) (blobStore blob_stores.BlobStoreInitialized, err error) {
	blobStore = blob_stores.MakeRemoteBlobStore(
		envLocal,
		cmd.GetBlobStoreConfigNamed(),
		envLocal.GetTempLocal(),
	)

	return
}
