package typed_blob_store

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type savedBlobFormatter struct {
	arf interfaces.BlobReaderFactory
}

func MakeSavedBlobFormatter(
	blobReaderFactory interfaces.BlobReaderFactory,
) savedBlobFormatter {
	return savedBlobFormatter{
		arf: blobReaderFactory,
	}
}

func (f savedBlobFormatter) FormatSavedBlob(
	w io.Writer,
	sh interfaces.MarklId,
) (n int64, err error) {
	var ar interfaces.BlobReader

	if ar, err = f.arf.MakeBlobReader(sh); err != nil {
		if errors.IsNotExist(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return n, err
	}

	defer errors.DeferredCloser(&err, ar)

	if n, err = io.Copy(w, ar); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}
