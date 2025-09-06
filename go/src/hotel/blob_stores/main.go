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
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"golang.org/x/crypto/ssh"
)

var defaultBuckets = []int{2}

const defaultHashTypeId = markl.HashTypeIdSha256

var (
	_ interfaces.BlobStore = localHashBucketed{}
	_ interfaces.BlobStore = &remoteSftp{}
)

type BlobStoreConfigNamed struct {
	Name     string
	BasePath string
	blob_store_configs.Config
}

type BlobStoreInitialized struct {
	BlobStoreConfigNamed
	interfaces.BlobStore
}

// TODO pass in custom UI context for printing
func MakeBlobStores(
	ctx interfaces.ActiveContext,
	envDir env_dir.Env,
	config genesis_configs.ConfigPrivate,
	directoryLayout interfaces.DirectoryLayout,
) (blobStores []BlobStoreInitialized) {
	if store_version.LessOrEqual(config.GetStoreVersion(), store_version.V10) {
		blobStores = make([]BlobStoreInitialized, 1)
		blob := config.(interfaces.BlobIOWrapperGetter)
		blobStores[0].Name = "0-default"
		blobStores[0].Config = blob.GetBlobIOWrapper().(blob_store_configs.Config)
		blobStores[0].BasePath = directoryLayout.DirBlobStores("blobs")
	} else {
		var configPaths []string

		{
			var err error

			if configPaths, err = files.DirNames(
				filepath.Join(directoryLayout.DirBlobStoreConfigs()),
			); err != nil {
				ctx.Cancel(err)
				return
			}
		}

		blobStores = make([]BlobStoreInitialized, len(configPaths))

		for i, configPath := range configPaths {
			blobStores[i].Name = fd.FileNameSansExt(configPath)
			blobStores[i].BasePath = directoryLayout.DirBlobStores(strconv.Itoa(i))

			if typedConfig, err := triple_hyphen_io.DecodeFromFile(
				blob_store_configs.Coder,
				configPath,
			); err != nil {
				ctx.Cancel(err)
				return
			} else {
				blobStores[i].Config = typedConfig.Blob
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
			return
		}
	}

	return
}

func MakeRemoteBlobStore(
	ctx interfaces.ActiveContext,
	config BlobStoreConfigNamed,
	tempFS env_dir.TemporaryFS,
	hashType markl.HashType,
) (blobStore BlobStoreInitialized) {
	blobStore.BlobStoreConfigNamed = config

	{
		var err error

		// TODO use sha of config to determine blob store base path
		if blobStore.BlobStore, err = MakeBlobStore(
			ctx,
			config,
			tempFS,
		); err != nil {
			ctx.Cancel(err)
			return
		}
	}

	return
}

// TODO describe base path agnostically
func MakeBlobStore(
	context interfaces.ActiveContext,
	config BlobStoreConfigNamed,
	tempFS env_dir.TemporaryFS,
) (store interfaces.BlobStore, err error) {
	printer := ui.MakePrefixPrinter(
		ui.Err(),
		fmt.Sprintf("(blob_store: %s) ", config.Name),
	)

	// TODO don't use tipe, use interfaces on the config
	switch tipe := config.GetBlobStoreType(); tipe {
	default:
		err = errors.BadRequestf("unsupported blob store type %q", tipe)
		return

	case "sftp":
		var sshClient *ssh.Client
		var configSFTP blob_store_configs.ConfigSFTPRemotePath

		switch config := config.Config.(type) {
		default:
			err = errors.BadRequestf("unsupported blob store config for type %q: %T", tipe, config)
			return

		case blob_store_configs.ConfigSFTPUri:
			if sshClient, err = MakeSSHClientFromSSHConfig(context, printer, config); err != nil {
				err = errors.Wrap(err)
				return
			}

			configSFTP = config

		case blob_store_configs.ConfigSFTPConfigExplicit:
			if sshClient, err = MakeSSHClientForExplicitConfig(context, printer, config); err != nil {
				err = errors.Wrap(err)
				return
			}

			configSFTP = config
		}

		return makeSftpStore(
			context,
			printer,
			configSFTP,
			sshClient,
		)

	case "local":
		if configLocal, ok := config.Config.(blob_store_configs.ConfigLocalHashBucketed); ok {
			return makeLocalHashBucketed(
				context,
				config.BasePath,
				// configLocal.GetBasePath(),
				configLocal,
				tempFS,
			)
		} else {
			err = errors.BadRequestf("unsupported blob store config for type %q: %T", tipe, config)
			return
		}
	}
}

func CopyBlobIfNecessary(
	env env_ui.Env,
	dst interfaces.BlobStore,
	src interfaces.BlobStore,
	blobId interfaces.MarklId,
	extraWriter io.Writer,
) (n int64, err error) {
	if src == nil {
		return
	}

	if err = markl.AssertIdIsNotNull(blobId, ""); err != nil {
		return
	}

	if dst.HasBlob(blobId) {
		err = env_dir.MakeErrBlobAlreadyExists(
			blobId,
			"",
		)

		return
	}

	if n, err = CopyBlob(env, dst, src, blobId, extraWriter); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO make this honor context closure and abort early
func CopyBlob(
	env env_ui.Env,
	dst interfaces.BlobStore,
	src interfaces.BlobStore,
	expectedDigest interfaces.MarklId,
	extraWriter io.Writer,
) (n int64, err error) {
	if src == nil {
		return
	}

	errors.PanicIfError(markl.AssertIdIsNotNull(expectedDigest, ""))

	var readCloser interfaces.ReadCloseMarklIdGetter

	if readCloser, err = src.BlobReader(expectedDigest); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.ContextMustClose(env, readCloser)

	var writeCloser interfaces.WriteCloseMarklIdGetter

	if writeCloser, err = dst.BlobWriter(""); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO should this be closed with an error when the shas don't match to
	// prevent a garbage object in the store?
	defer errors.ContextMustClose(env, writeCloser)

	outputWriter := io.Writer(writeCloser)

	if extraWriter != nil {
		outputWriter = io.MultiWriter(outputWriter, extraWriter)
	}

	if n, err = io.Copy(outputWriter, readCloser); err != nil {
		err = errors.Wrap(err)
		return
	}

	readerDigest := readCloser.GetMarklId()
	writerDigest := writeCloser.GetMarklId()

	if !markl.Equals(readerDigest, expectedDigest) ||
		!markl.Equals(writerDigest, expectedDigest) {
		err = errors.Errorf(
			"lookup sha was %s, read sha was %s, but written sha was %s",
			expectedDigest,
			readerDigest,
			writerDigest,
		)
	}

	return
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

	var readCloser interfaces.ReadCloseMarklIdGetter

	if readCloser, err = blobStore.BlobReader(expected); err != nil {
		err = errors.Wrap(err)
		return
	}

	if _, err = io.Copy(progressWriter, readCloser); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = markl.MakeErrNotEqual(
		expected,
		readCloser.GetMarklId(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = readCloser.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
