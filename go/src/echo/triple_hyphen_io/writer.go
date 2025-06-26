package triple_hyphen_io

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type Writer struct {
	Metadata, Blob io.WriterTo
}

func (w1 Writer) WriteTo(w2 io.Writer) (n int64, err error) {
	w := bufio.NewWriter(w2)
	defer errors.DeferredFlusher(&err, w)

	var n1 int64
	var n2 int

	hasMetadataContent := true

	if mwt, ok := w1.Metadata.(MetadataWriterTo); ok {
		hasMetadataContent = mwt.HasMetadataContent()
	}

	if w1.Metadata != nil && hasMetadataContent {
		n2, err = w.WriteString(Boundary + "\n")
		n += int64(n2)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		n1, err = w1.Metadata.WriteTo(w)
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		w.WriteString(Boundary + "\n")
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		if w1.Blob != nil {
			w.WriteString("\n")
			n += n1

			if err != nil {
				err = errors.Wrap(err)
				return
			}
		}
	}

	if w1.Blob != nil {
		n1, err = w1.Blob.WriteTo(w)
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
