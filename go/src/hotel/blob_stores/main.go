package blob_stores

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blob_ids"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"golang.org/x/crypto/ssh"
)

var defaultBuckets = []int{2}

type BlobStoreInitialized struct {
	Name     string
	BasePath string
	blob_store_configs.Config
	interfaces.BlobStore
}

// TODO describe base path agnostically
func MakeBlobStore(
	ctx interfaces.ActiveContext,
	basePath string,
	config blob_store_configs.Config,
	tempFS env_dir.TemporaryFS,
) (store interfaces.BlobStore, err error) {
	switch tipe := config.GetBlobStoreType(); tipe {
	default:
		err = errors.BadRequestf("unsupported blob store type %q", tipe)
		return

	case "sftp":
		var sshClient *ssh.Client
		var configSFTP blob_store_configs.ConfigSFTPRemotePath

		switch config := config.(type) {
		default:
			err = errors.BadRequestf("unsupported blob store config for type %q: %T", tipe, config)
			return

		case blob_store_configs.ConfigSFTPUri:
			if sshClient, err = MakeSSHClientFromSSHConfig(ctx, config); err != nil {
				err = errors.Wrap(err)
				return
			}

			configSFTP = config

		case blob_store_configs.ConfigSFTPConfigExplicit:
			if sshClient, err = MakeSSHClientForExplicitConfig(ctx, config); err != nil {
				err = errors.Wrap(err)
				return
			}

			configSFTP = config
		}

		return makeSftpStore(ctx, configSFTP, sshClient)

	case "local":
		if config, ok := config.(blob_store_configs.ConfigLocalHashBucketed); ok {
			return makeLocalHashBucketed(ctx, basePath, config, tempFS)
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
	blobShaGetter interfaces.BlobIdGetter,
	extraWriter io.Writer,
) (n int64, err error) {
	if src == nil {
		return
	}

	blobSha := blobShaGetter.GetBlobId()

	if dst.HasBlob(blobSha) || blobSha.IsNull() {
		err = env_dir.MakeErrBlobAlreadyExists(
			blobSha,
			"",
		)

		return
	}

	return CopyBlob(env, dst, src, blobSha, extraWriter)
}

// TODO make this honor context closure and abort early
func CopyBlob(
	env env_ui.Env,
	dst interfaces.BlobStore,
	src interfaces.BlobStore,
	blobSha interfaces.BlobId,
	extraWriter io.Writer,
) (n int64, err error) {
	if src == nil {
		return
	}

	var rc interfaces.ReadCloseBlobIdGetter

	if rc, err = src.BlobReader(blobSha); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.ContextMustClose(env, rc)

	var wc interfaces.WriteCloseBlobIdGetter

	if wc, err = dst.BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO should this be closed with an error when the shas don't match to
	// prevent a garbage object in the store?
	defer errors.ContextMustClose(env, wc)

	outputWriter := io.Writer(wc)

	if extraWriter != nil {
		outputWriter = io.MultiWriter(outputWriter, extraWriter)
	}

	if n, err = io.Copy(outputWriter, rc); err != nil {
		err = errors.Wrap(err)
		return
	}

	shaRc := rc.GetBlobId()
	shaWc := wc.GetBlobId()

	if !blob_ids.Equals(shaRc, blobSha) ||
		!blob_ids.Equals(shaWc, blobSha) {
		err = errors.ErrorWithStackf(
			"lookup sha was %s, read sha was %s, but written sha was %s",
			blobSha,
			shaRc,
			shaWc,
		)
	}

	return
}

// TODO offer options like just checking the existence of the blob, getting its
// size, or full verification
func VerifyBlob(
	ctx interfaces.Context,
	blobStore interfaces.BlobStore,
	sh interfaces.BlobId,
	progressWriter io.Writer,
) (err error) {
	// TODO check if `blobStore` implements a `VerifyBlob` method and call that
	// instead (for expensive blob stores that may implement their own remote
	// verification, such as ssh, sftp, or something else)

	var readCloser interfaces.ReadCloseBlobIdGetter

	if readCloser, err = blobStore.BlobReader(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	if _, err = io.Copy(progressWriter, readCloser); err != nil {
		err = errors.Wrap(err)
		return
	}

	expected := sha.MustWithDigest(sh)

	if err = expected.AssertEqualsShaLike(readCloser.GetBlobId()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = readCloser.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
