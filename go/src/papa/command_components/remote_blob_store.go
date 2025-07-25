package command_components

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
)

// TODO migrate to using `command_components.BlobStore`
type RemoteBlobStore struct {
	Blobs  string
	Config blob_store_configs.TomlV0
}

func (cmd *RemoteBlobStore) SetFlagSet(flagSet *flag.FlagSet) {
	cmd.Config.CompressionType = compression_type.CompressionTypeDefault
	cmd.Config.CompressionType.SetFlagSet(flagSet)
	flagSet.StringVar(&cmd.Blobs, "blobs", "", "")
}

func (cmd *RemoteBlobStore) MakeRemoteBlobStore(
	envLocal env_local.Env,
) (blobStore interfaces.BlobStore, err error) {
	if blobStore, err = blob_stores.MakeBlobStore(
		envLocal,
		cmd.Blobs,
		&cmd.Config,
		envLocal.GetTempLocal(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
