package blob_library

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type format[
	BLOB any,
	BLOB_PTR interfaces.Ptr[BLOB],
] struct {
	interfaces.DecoderFromReader[BLOB_PTR]
	interfaces.SavedBlobFormatter
	interfaces.EncoderToWriter[BLOB_PTR]
}

func MakeBlobFormat[
	BLOB any,
	BLOB_PTR interfaces.Ptr[BLOB],
](
	decoder interfaces.DecoderFromReader[BLOB_PTR],
	encoder interfaces.EncoderToWriter[BLOB_PTR],
	blobReader interfaces.BlobReaderFactory,
) interfaces.Format[BLOB, BLOB_PTR] {
	return format[BLOB, BLOB_PTR]{
		DecoderFromReader:  decoder,
		EncoderToWriter:    encoder,
		SavedBlobFormatter: MakeSavedBlobFormatter(blobReader),
	}
}

func (af format[BLOB, BLOB_PTR]) EncodeTo(
	object BLOB_PTR,
	writer io.Writer,
) (n int64, err error) {
	if af.EncoderToWriter == nil {
		err = errors.ErrorWithStackf("no ParsedBlobFormatter")
	} else {
		n, err = af.EncoderToWriter.EncodeTo(object, writer)
	}

	return n, err
}
