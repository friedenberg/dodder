package blob_stores

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
)

// TODO offer options like just checking the existence of the blob, getting its
// size, or full verification
func VerifyBlob(
	ctx errors.Context,
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
