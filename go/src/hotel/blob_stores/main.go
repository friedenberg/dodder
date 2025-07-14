package blob_stores

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
)

func MakeBlobStore(
	ctx errors.Context,
	basePath string,
	config blob_store_configs.Config,
	tempFS env_dir.TemporaryFS,
) (store interfaces.LocalBlobStore, err error) {
	switch config := config.(type) {
	default:
		err = errors.BadRequestf("unsupported blob store config: %T", config)
		return

	case sftpConfig:
		return makeSftpStore(ctx, config)

	case blob_store_configs.ConfigLocalGitLikeBucketed:
		return makeGitLikeBucketedStore(ctx, basePath, config, tempFS)
	}
}

func CopyBlobIfNecessary(
	env env_ui.Env,
	dst interfaces.BlobStore,
	src interfaces.BlobStore,
	blobShaGetter interfaces.ShaGetter,
	extraWriter io.Writer,
) (n int64, err error) {
	if src == nil {
		return
	}

	blobSha := blobShaGetter.GetShaLike()

	if dst.HasBlob(blobSha) || blobSha.IsNull() {
		err = env_dir.MakeErrAlreadyExists(
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
	blobSha interfaces.Sha,
	extraWriter io.Writer,
) (n int64, err error) {
	if src == nil {
		return
	}

	var rc sha.ReadCloser

	if rc, err = src.BlobReader(blobSha); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer env.MustClose(rc)

	var wc sha.WriteCloser

	if wc, err = dst.BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO should this be closed with an error when the shas don't match to
	// prevent a garbage object in the store?
	defer env.MustClose(wc)

	outputWriter := io.Writer(wc)

	if extraWriter != nil {
		outputWriter = io.MultiWriter(outputWriter, extraWriter)
	}

	if n, err = io.Copy(outputWriter, rc); err != nil {
		err = errors.Wrap(err)
		return
	}

	shaRc := rc.GetShaLike()
	shaWc := wc.GetShaLike()

	if !shaRc.EqualsSha(blobSha) || !shaWc.EqualsSha(blobSha) {
		err = errors.ErrorWithStackf(
			"lookup sha was %s, read sha was %s, but written sha was %s",
			blobSha,
			shaRc,
			shaWc,
		)
	}

	return
}

func VerifyBlob(
	ctx errors.Context,
	blobStore interfaces.LocalBlobStore,
	sh interfaces.Sha,
	progressWriter io.Writer,
) (err error) {
	// TODO check if `blobStore` implements a `VerifyBlob` method and call that
	// instead (for expensive blob stores that may implement their own remote
	// verification, such as ssh, sftp, or something else)

	var readCloser interfaces.ShaReadCloser

	if readCloser, err = blobStore.BlobReader(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	if _, err = io.Copy(progressWriter, readCloser); err != nil {
		err = errors.Wrap(err)
		return
	}

	expected := sha.Make(sh)

	if err = expected.AssertEqualsShaLike(readCloser.GetShaLike()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = readCloser.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
