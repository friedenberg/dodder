package typed_blob_store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type format2[
	O any,
] struct {
	interfaces.DecoderFromReader[O]
	interfaces.SavedBlobFormatter
	interfaces.EncoderToWriter[O]
}

func MakeBlobFormat2[
	O any,
](
	decoder interfaces.DecoderFromReader[O],
	encoder interfaces.EncoderToWriter[O],
	blobReader interfaces.BlobReader,
) interfaces.Format[O] {
	return format2[O]{
		DecoderFromReader:  decoder,
		EncoderToWriter:    encoder,
		SavedBlobFormatter: MakeSavedBlobFormatter(blobReader),
	}
}
