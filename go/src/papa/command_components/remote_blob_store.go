package command_components

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_config"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_store"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
)

type RemoteBlobStore struct {
	Blobs  string
	Config blob_store_config.TomlV0
}

func (cmd *RemoteBlobStore) SetFlagSet(f *flag.FlagSet) {
	cmd.Config.CompressionType = compression_type.CompressionTypeDefault
	cmd.Config.CompressionType.SetFlagSet(f)
	f.StringVar(&cmd.Blobs, "blobs", "", "")
}

func (cmd *RemoteBlobStore) MakeRemoteBlobStore(
	envLocal env_local.Env,
) (blobStore interfaces.BlobStore, err error) {
	blobStore = blob_store.MakeBlobStore(
		envLocal,
		cmd.Blobs,
		&cmd.Config,
		envLocal.GetTempLocal(),
	)

	return
}
