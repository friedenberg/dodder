package blob_stores

import (
	"fmt"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/directory_layout"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"golang.org/x/crypto/ssh"
)

var defaultBuckets = []int{2}

// TODO pass in custom UI context for printing
// TODO consolidated envDir and ctx arguments
func MakeBlobStoresFromRepoConfig(
	ctx interfaces.ActiveContext,
	envDir env_dir.Env,
	config genesis_configs.ConfigPrivate,
	directoryLayout directory_layout.BlobStore,
) (blobStores []BlobStoreInitialized) {
	if store_version.LessOrEqual(config.GetStoreVersion(), store_version.V10) {
		blobStores = make([]BlobStoreInitialized, 1)
		blob := config.(interfaces.BlobIOWrapperGetter)
		blobStores[0].Config.Blob = blob.GetBlobIOWrapper().(blob_store_configs.Config)
		blobStores[0].Config.Type = ids.GetOrPanic(
			ids.TypeTomlBlobStoreConfigV0,
		).Type
		blobStores[0].Path = directory_layout.BlobStorePath{
			Base:   directory_layout.DirBlobStore(directoryLayout, "blobs"),
			Config: "0-default",
		}
	} else {
		configPaths := directory_layout.GetBlobStoreConfigPaths(ctx, directoryLayout)
		blobStores = make([]BlobStoreInitialized, len(configPaths))

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
	}

	for index, blobStore := range blobStores {
		var err error

		// TODO use sha of config to determine blob store base path
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

func MakeBlobStores(
	ctx interfaces.ActiveContext,
	envDir env_dir.Env,
	directoryLayout directory_layout.BlobStore,
) (blobStores []BlobStoreInitialized) {
	configPaths := directory_layout.GetBlobStoreConfigPaths(
		ctx,
		directoryLayout,
	)

	blobStores = make([]BlobStoreInitialized, len(configPaths))

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

	for index, blobStore := range blobStores {
		var err error

		// TODO use sha of config to determine blob store base path
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

		// TODO use sha of config to determine blob store base path
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
		fmt.Sprintf("(blob_store: %s) ", configNamed.Path.Config),
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
			configNamed.Path.Base,
			config,
		)

	case blob_store_configs.ConfigPointer:
		var typedConfig triple_hyphen_io.TypedBlob[blob_store_configs.Config]

		if typedConfig, err = triple_hyphen_io.DecodeFromFile(
			blob_store_configs.Coder,
			config.GetConfigPath(),
		); err != nil {
			err = errors.Wrap(err)
			return store, err
		}

		configNamed.Config = typedConfig
		configNamed.Path = directory_layout.BlobStorePath{
			Base:   filepath.Dir(config.GetConfigPath()),
			Config: config.GetConfigPath(),
		}

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
