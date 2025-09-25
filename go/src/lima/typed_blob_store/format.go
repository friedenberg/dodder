package typed_blob_store

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type format[
	O any,
	OPtr interfaces.Ptr[O],
] struct {
	interfaces.DecoderFromReader[OPtr]
	interfaces.SavedBlobFormatter
	interfaces.EncoderToWriter[OPtr]
}

func MakeBlobFormat[
	O any,
	OPtr interfaces.Ptr[O],
](
	decoder interfaces.DecoderFromReader[OPtr],
	encoder interfaces.EncoderToWriter[OPtr],
	blobReader interfaces.BlobReaderFactory,
) Format[O, OPtr] {
	return format[O, OPtr]{
		DecoderFromReader:  decoder,
		EncoderToWriter:    encoder,
		SavedBlobFormatter: MakeSavedBlobFormatter(blobReader),
	}
}

func (af format[O, OPtr]) EncodeTo(
	object OPtr,
	writer io.Writer,
) (n int64, err error) {
	if af.EncoderToWriter == nil {
		err = errors.ErrorWithStackf("no ParsedBlobFormatter")
	} else {
		n, err = af.EncoderToWriter.EncodeTo(object, writer)
	}

	return n, err
}
