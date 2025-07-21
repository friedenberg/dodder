package sha

import (
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func DebugCompareBlobs(
	ctx interfaces.Context,
	blobStore interfaces.BlobStore,
	expectedSha interfaces.Digest,
	actual *strings.Builder,
) {
	var expected strings.Builder

	var blobReader interfaces.ShaReadCloser

	{
		var err error

		if blobReader, err = blobStore.BlobReader(expectedSha); err != nil {
			ctx.Cancel(err)
		}
	}

	if _, err := io.Copy(&expected, blobReader); err != nil {
		ctx.Cancel(err)
	}

	ui.Debug().Printf("expected: %q, actual: %q", &expected, actual)
}
