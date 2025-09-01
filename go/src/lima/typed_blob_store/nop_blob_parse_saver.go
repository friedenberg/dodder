package typed_blob_store

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type nopBlobParseSaver[
	OBJECT any,
	OBJECT_PTR interfaces.Ptr[OBJECT],
] struct {
	blobStore interfaces.BlobWriter
}

func MakeNopBlobParseSaver[
	OBJECT any,
	OBJECT_PTR interfaces.Ptr[OBJECT],
](awf interfaces.BlobWriter,
) nopBlobParseSaver[OBJECT, OBJECT_PTR] {
	return nopBlobParseSaver[OBJECT, OBJECT_PTR]{
		blobStore: awf,
	}
}

func (parseSaver nopBlobParseSaver[OBJECT, OBJECT_PTR]) ParseBlob(
	reader io.Reader,
	object OBJECT_PTR,
) (n int64, err error) {
	var blobWriter interfaces.WriteCloseMarklIdGetter

	if blobWriter, err = parseSaver.blobStore.BlobWriter(""); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, blobWriter)

	if n, err = io.Copy(blobWriter, reader); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
