package triple_hyphen_io

import (
	"bufio"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
)

type Reader struct {
	RequireMetadata bool // TODO-P4 add delimiter
	Metadata, Blob  io.ReaderFrom
}

func (mr *Reader) ReadFrom(r io.Reader) (n int64, err error) {
	var n1 int64
	n1, err = mr.readMetadataFrom(&r)
	n += n1

	if err != nil {
		err = errors.Wrapf(err, "metadata read failed")
		return
	}

	n1, err = mr.Blob.ReadFrom(bufio.NewReader(r))
	n += n1

	if err != nil {
		err = errors.Wrapf(err, "blob read failed")
		return
	}

	return
}

func (mr *Reader) readMetadataFrom(r *io.Reader) (n int64, err error) {
	var state readerState
	br := bufio.NewReader(*r)

	if mr.RequireMetadata && mr.Metadata == nil {
		err = errors.ErrorWithStackf("metadata reader is nil")
		return
	}

	if mr.Blob == nil {
		err = errors.ErrorWithStackf("blob reader is nil")
		return
	}

	var object_metadata ohio.PipedReader

	isEOF := false

LINE_READ_LOOP:
	for !isEOF {
		var rawLine, line string

		rawLine, err = br.ReadString('\n')
		n += int64(len(rawLine))

		if err != nil && err != io.EOF {
			err = errors.Wrap(err)
			return
		}

		if errors.IsEOF(err) {
			err = nil
			isEOF = true
		}

		line = strings.TrimSuffix(rawLine, "\n")

		switch state {
		case readerStateEmpty:
			switch {
			case mr.RequireMetadata && line != Boundary:
				err = errors.ErrorWithStackf("expected %q but got %q", Boundary, line)
				return

			case line != Boundary:
				*r = io.MultiReader(
					strings.NewReader(rawLine),
					br,
				)

				break LINE_READ_LOOP
			}

			state += 1

			object_metadata = ohio.MakePipedReaderFrom(mr.Metadata)

		case readerStateFirstBoundary:
			if line == Boundary {
				if _, err = object_metadata.Close(); err != nil {
					err = errors.Wrapf(err, "metadata read failed")
					return
				}

				state += 1
				break
			}

			if _, err = object_metadata.Write([]byte(rawLine)); err != nil {
				err = errors.Wrap(err)
				return
			}

		case readerStateSecondBoundary:
			*r = br
			break LINE_READ_LOOP

		default:
			err = errors.ErrorWithStackf("impossible state %d", state)
			return
		}
	}

	return
}
