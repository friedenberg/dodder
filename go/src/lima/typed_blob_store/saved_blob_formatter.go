package typed_blob_store

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type savedBlobFormatter struct {
	arf interfaces.BlobReader
}

func MakeSavedBlobFormatter(
	blobReaderFactory interfaces.BlobReader,
) savedBlobFormatter {
	return savedBlobFormatter{
		arf: blobReaderFactory,
	}
}

func (f savedBlobFormatter) FormatSavedBlob(
	w io.Writer,
	sh interfaces.MarklId,
) (n int64, err error) {
	var ar interfaces.ReadCloseMarklIdGetter

	if ar, err = f.arf.BlobReader(sh); err != nil {
		if errors.IsNotExist(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	defer errors.DeferredCloser(&err, ar)

	if n, err = io.Copy(w, ar); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
