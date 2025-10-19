package blob_stores

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"golang.org/x/crypto/ssh"
)

var defaultBuckets = []int{2}

type BlobStoreConfigNamed struct {
	Name     string
	BasePath string
	Config   blob_store_configs.TypedConfig
}

type BlobStoreInitialized struct {
	BlobStoreConfigNamed
	interfaces.BlobStore
}

// TODO pass in custom UI context for printing
func MakeBlobStoresFromRepoConfig(
	ctx interfaces.ActiveContext,
	envDir env_dir.Env,
	config genesis_configs.ConfigPrivate,
	directoryLayout interfaces.BlobStoreDirectoryLayout,
) (blobStores []BlobStoreInitialized) {
	if store_version.LessOrEqual(config.GetStoreVersion(), store_version.V10) {
		blobStores = make([]BlobStoreInitialized, 1)
		blob := config.(interfaces.BlobIOWrapperGetter)
		blobStores[0].Name = "0-default"
		blobStores[0].Config.Blob = blob.GetBlobIOWrapper().(blob_store_configs.Config)
		blobStores[0].Config.Type = ids.GetOrPanic(
			ids.TypeTomlBlobStoreConfigV0,
		).Type
		blobStores[0].BasePath = interfaces.DirectoryLayoutDirBlobStore(
			directoryLayout,
			"blobs",
		)
	} else {
		var configPaths []string

		{
			var err error

			if configPaths, err = files.DirNames(
				filepath.Join(directoryLayout.DirBlobStoreConfigs()),
			); err != nil {
				ctx.Cancel(err)
				return blobStores
			}
		}

		blobStores = make([]BlobStoreInitialized, len(configPaths))

		for i, configPath := range configPaths {
			blobStores[i].Name = fd.FileNameSansExt(configPath)
			blobStores[i].BasePath = interfaces.DirectoryLayoutDirBlobStore(directoryLayout, strconv.Itoa(i))

			if typedConfig, err := triple_hyphen_io.DecodeFromFile(
				blob_store_configs.Coder,
				configPath,
			); err != nil {
				ctx.Cancel(err)
				return blobStores
			} else {
				blobStores[i].Config = typedConfig
			}
		}
	}

	for i, blobStore := range blobStores {
		var err error

		// TODO use sha of config to determine blob store base path
		if blobStores[i].BlobStore, err = MakeBlobStore(
			ctx,
			blobStore.BlobStoreConfigNamed,
			envDir.GetTempLocal(),
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
	directoryLayout interfaces.BlobStoreDirectoryLayout,
) (blobStores []BlobStoreInitialized) {
	var configPaths []string

	{
		var err error

		if configPaths, err = files.DirNames(
			filepath.Join(directoryLayout.DirBlobStoreConfigs()),
		); err != nil {
			if errors.IsNotExist(err) {
				return blobStores
			} else {
				ctx.Cancel(err)
				// no-op
				return blobStores
			}
		}
	}

	blobStores = make([]BlobStoreInitialized, len(configPaths))

	for i, configPath := range configPaths {
		blobStores[i].Name = fd.FileNameSansExt(configPath)
		blobStores[i].BasePath = interfaces.DirectoryLayoutDirBlobStore(
			directoryLayout,
			strconv.Itoa(i),
		)

		if typedConfig, err := triple_hyphen_io.DecodeFromFile(
			blob_store_configs.Coder,
			configPath,
		); err != nil {
			ctx.Cancel(err)
			return blobStores
		} else {
			blobStores[i].Config = typedConfig
		}
	}

	for i, blobStore := range blobStores {
		var err error

		// TODO use sha of config to determine blob store base path
		if blobStores[i].BlobStore, err = MakeBlobStore(
			ctx,
			blobStore.BlobStoreConfigNamed,
			envDir.GetTempLocal(),
		); err != nil {
			ctx.Cancel(err)
			return blobStores
		}
	}

	return blobStores
}

func MakeRemoteBlobStore(
	ctx interfaces.ActiveContext,
	configNamed BlobStoreConfigNamed,
	tempFS env_dir.TemporaryFS,
) (blobStore BlobStoreInitialized) {
	blobStore.BlobStoreConfigNamed = configNamed

	{
		var err error

		// TODO use sha of config to determine blob store base path
		if blobStore.BlobStore, err = MakeBlobStore(
			ctx,
			configNamed,
			tempFS,
		); err != nil {
			ctx.Cancel(err)
			return blobStore
		}
	}

	return blobStore
}

// TODO describe base path agnostically
func MakeBlobStore(
	context interfaces.ActiveContext,
	configNamed BlobStoreConfigNamed,
	tempFS env_dir.TemporaryFS,
) (store interfaces.BlobStore, err error) {
	printer := ui.MakePrefixPrinter(
		ui.Err(),
		fmt.Sprintf("(blob_store: %s) ", configNamed.Name),
	)

	// TODO don't use tipe, use interfaces on the config
	// switch tipe := config.Config.Blob.GetBlobStoreType(); tipe {
	configBlob := configNamed.Config.Blob

	switch config := configBlob.(type) {
	case blob_store_configs.ConfigSFTPUri:
		return makeSftpStore(
			context,
			printer,
			config,
			func() (*ssh.Client, error) {
				return MakeSSHClientFromSSHConfig(context, printer, config)
			},
		)

	case blob_store_configs.ConfigSFTPConfigExplicit:
		return makeSftpStore(
			context,
			printer,
			config,
			func() (*ssh.Client, error) {
				return MakeSSHClientForExplicitConfig(context, printer, config)
			},
		)

	case blob_store_configs.ConfigLocalHashBucketed:
		return makeLocalHashBucketed(
			context,
			configNamed.BasePath,
			config,
			tempFS,
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

		return MakeBlobStore(context, configNamed, tempFS)

	default:
		err = errors.BadRequestf(
			"unsupported blob store type %q:%T",
			configBlob.GetBlobStoreType(),
			configBlob,
		)

		return store, err
	}
}

// TODO offer options like just checking the existence of the blob, getting its
// size, or full verification
func VerifyBlob(
	ctx interfaces.Context,
	blobStore interfaces.BlobStore,
	expected interfaces.MarklId,
	progressWriter io.Writer,
) (err error) {
	// TODO check if `blobStore` implements a `VerifyBlob` method and call that
	// instead (for expensive blob stores that may implement their own remote
	// verification, such as ssh, sftp, or something else)

	var readCloser interfaces.BlobReader

	if readCloser, err = blobStore.MakeBlobReader(expected); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if _, err = io.Copy(progressWriter, readCloser); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = markl.AssertEqual(
		expected,
		readCloser.GetMarklId(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = readCloser.Close(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
