package inventory_list_blobs

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func makeScanner(
	in io.Reader,
	of object_inventory_format.Format,
	op object_inventory_format.Options,
) *scanner {
	return &scanner{
		ringBuffer: catgut.MakeRingBuffer(in, 0),
		format:     of,
		options:    op,
		es:         ids.MakeTagMutableSet(),
	}
}

type scanner struct {
	object_probe_index.Range

	ringBuffer *catgut.RingBuffer
	format     object_inventory_format.Format
	options    object_inventory_format.Options
	afterFirst bool

	m  object_metadata.Metadata
	g  genres.Genre
	es ids.TagMutableSet
	k  string

	err     error
	lastSku *sku.Transacted
	debug   bool
}

func (scanner *scanner) SetDebug() {
	scanner.debug = true
}

func (scanner *scanner) Error() error {
	if errors.IsEOF(scanner.err) {
		return nil
	}

	return scanner.err
}

func (scanner *scanner) GetTransacted() *sku.Transacted {
	return scanner.lastSku
}

func (scanner *scanner) GetRange() object_probe_index.Range {
	return scanner.Range
}

func (scanner *scanner) Scan() (ok bool) {
	if scanner.err != nil {
		return
	}

	var n1 int64

	scanner.lastSku = nil

	if !scanner.afterFirst {
		scanner.err = triple_hyphen_io.ReadBoundaryFromPeeker(scanner.ringBuffer)

		if errors.IsEOF(scanner.err) {
			return
		} else if scanner.err != nil {
			scanner.err = errors.Wrap(scanner.err)
			return
		}

		scanner.afterFirst = true
	}

	scanner.Offset += int64(len(triple_hyphen_io.Boundary) + 1)
	scanner.ContentLength = 0

	scanner.lastSku = sku.GetTransactedPool().Get()

	scanner.ContentLength, scanner.err = scanner.format.ParsePersistentMetadata(
		scanner.ringBuffer,
		scanner.lastSku,
		scanner.options,
	)

	if scanner.ContentLength == 0 {
		if scanner.err == io.EOF {
			return
		} else if scanner.err != nil {
			scanner.err = errors.Wrapf(scanner.err, "Bytes: %d", n1)
			scanner.err = errors.Wrapf(scanner.err, "Holder: %v", scanner.lastSku)
			return
		}
	}

	oldErr := scanner.err

	scanner.err = triple_hyphen_io.ReadBoundaryFromPeeker(scanner.ringBuffer)

	if errors.IsNotNilAndNotEOF(scanner.err) {
		scanner.err = errors.Wrap(errors.MakeMulti(scanner.err, oldErr))
		return
	} else if scanner.err == io.EOF {
		scanner.err = nil
		return
	}

	ok = true

	return
}
