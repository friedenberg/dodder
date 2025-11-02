package blob_stores

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blob_store_id"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/directory_layout"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"golang.org/x/crypto/ssh"
)

var defaultBuckets = []int{2}

func makeBlobStoreConfigs(
	ctx interfaces.ActiveContext,
	directoryLayout directory_layout.BlobStore,
) []BlobStoreInitialized {
	configPaths := directory_layout.GetBlobStoreConfigPaths(
		ctx,
		directoryLayout,
	)

	blobStores := make([]BlobStoreInitialized, len(configPaths))

	for index, configPath := range configPaths {
		blobStores[index].Path = directory_layout.GetBlobStorePath(
			directoryLayout,
			fd.DirBaseOnly(configPath),
		)

		if typedConfig, err := triple_hyphen_io.DecodeFromFile(
			blob_store_configs.Coder,
			configPath,
		); err != nil {
			ctx.Cancel(err)
			return blobStores
		} else {
			blobStores[index].Config = typedConfig
		}
	}

	return blobStores
}

// TODO pass in custom UI context for printing
// TODO consolidated envDir and ctx arguments
func MakeBlobStores(
	ctx interfaces.ActiveContext,
	envDir env_dir.Env,
	directoryLayout directory_layout.BlobStore,
) (blobStores []BlobStoreInitialized) {
	// based on explicit xdg (that is, may include override)
	blobStores = makeBlobStoreConfigs(ctx, directoryLayout)

	if envDir.GetXDG().GetLocationType() == blob_store_id.LocationTypeOverride {
		if directoryLayoutForXDG, err := directory_layout.CloneBlobStoreWithXDG(
			directoryLayout,
			envDir.GetXDG().CloneWithoutOverride(),
		); err != nil {
			ctx.Cancel(err)
			return blobStores
		} else {
			blobStoresForXDG := makeBlobStoreConfigs(ctx, directoryLayoutForXDG)
			blobStores = append(blobStores, blobStoresForXDG...)
		}
	}

	for index, blobStore := range blobStores {
		var err error

		if blobStores[index].BlobStore, err = MakeBlobStore(
			envDir,
			blobStore.ConfigNamed,
		); err != nil {
			ctx.Cancel(err)
			return blobStores
		}
	}

	return blobStores
}

func MakeRemoteBlobStore(
	envDir env_dir.Env,
	configNamed blob_store_configs.ConfigNamed,
) (blobStore BlobStoreInitialized) {
	blobStore.ConfigNamed = configNamed

	{
		var err error

		if blobStore.BlobStore, err = MakeBlobStore(
			envDir,
			configNamed,
		); err != nil {
			envDir.GetActiveContext().Cancel(err)
			return blobStore
		}
	}

	return blobStore
}

// TODO describe base path agnostically
func MakeBlobStore(
	envDir env_dir.Env,
	configNamed blob_store_configs.ConfigNamed,
) (store interfaces.BlobStore, err error) {
	printer := ui.MakePrefixPrinter(
		ui.Err(),
		fmt.Sprintf("(blob_store: %s) ", configNamed.Path.GetConfig()),
	)

	configBlob := configNamed.Config.Blob

	switch config := configBlob.(type) {
	case blob_store_configs.ConfigSFTPUri:
		return makeSftpStore(
			envDir.GetActiveContext(),
			printer,
			config,
			func() (*ssh.Client, error) {
				return MakeSSHClientFromSSHConfig(
					envDir.GetActiveContext(),
					printer,
					config,
				)
			},
		)

	case blob_store_configs.ConfigSFTPConfigExplicit:
		return makeSftpStore(
			envDir.GetActiveContext(),
			printer,
			config,
			func() (*ssh.Client, error) {
				return MakeSSHClientForExplicitConfig(
					envDir.GetActiveContext(),
					printer,
					config,
				)
			},
		)

	case blob_store_configs.ConfigLocalHashBucketed:
		return makeLocalHashBucketed(
			envDir,
			configNamed.Path.GetBase(),
			config,
		)

	case blob_store_configs.ConfigPointer:
		var typedConfig triple_hyphen_io.TypedBlob[blob_store_configs.Config]
		otherStorePath := config.GetPath()

		if typedConfig, err = triple_hyphen_io.DecodeFromFile(
			blob_store_configs.Coder,
			otherStorePath.GetConfig(),
		); err != nil {
			err = errors.Wrap(err)
			return store, err
		}

		configNamed.Config = typedConfig
		configNamed.Path = directory_layout.MakeBlobStorePath(
			configNamed.Path.GetId(),
			otherStorePath.GetBase(),
			otherStorePath.GetConfig(),
		)

		return MakeBlobStore(envDir, configNamed)

	default:
		err = errors.BadRequestf(
			"unsupported blob store type %q:%T",
			configBlob.GetBlobStoreType(),
			configBlob,
		)

		return store, err
	}
}
