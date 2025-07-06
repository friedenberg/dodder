package triple_hyphen_io2

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type Writer struct {
	Metadata, Blob io.WriterTo
}

func (writer Writer) WriteTo(ioWriter io.Writer) (n int64, err error) {
	bufferedWriter := bufio.NewWriter(ioWriter)
	defer errors.DeferredFlusher(&err, bufferedWriter)

	var n1 int64
	var n2 int

	hasMetadataContent := true

	if mwt, ok := writer.Metadata.(MetadataWriterTo); ok {
		hasMetadataContent = mwt.HasMetadataContent()
	}

	if writer.Metadata != nil && hasMetadataContent {
		n2, err = bufferedWriter.WriteString(Boundary + "\n")
		n += int64(n2)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		n1, err = writer.Metadata.WriteTo(bufferedWriter)
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		bufferedWriter.WriteString(Boundary + "\n")
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		if writer.Blob != nil {
			bufferedWriter.WriteString("\n")
			n += n1

			if err != nil {
				err = errors.Wrap(err)
				return
			}
		}
	}

	if writer.Blob != nil {
		n1, err = writer.Blob.WriteTo(bufferedWriter)
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
