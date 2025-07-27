package triple_hyphen_io

import (
	"bufio"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
)

type Decoder[BLOB any] struct {
	RequireMetadata bool
	Metadata, Blob  interfaces.DecoderFromBufferedReader[BLOB]
}

func (decoder *Decoder[BLOB]) DecodeFrom(
	blob BLOB,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	var n1 int64
	n1, err = decoder.readMetadataFrom(blob, bufferedReader)
	n += n1

	if err != nil {
		err = errors.Wrapf(err, "metadata read failed")
		return
	}

	n1, err = decoder.Blob.DecodeFrom(blob, bufferedReader)
	n += n1

	if err != nil {
		err = errors.Wrapf(err, "blob read failed")
		return
	}

	return
}

func (decoder *Decoder[BLOB]) readMetadataFrom(
	blob BLOB,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	var state readerState

	if decoder.RequireMetadata && decoder.Metadata == nil {
		err = errors.ErrorWithStackf("metadata reader is nil")
		return
	}

	if decoder.Blob == nil {
		err = errors.ErrorWithStackf("blob reader is nil")
		return
	}

	var metadataPipe ohio.PipedReader

	{
		var isBoundary bool

		if err = ReadBoundaryFromPeeker(bufferedReader); err != nil {
			if err == errBoundaryInvalid {
				err = nil
			} else {
				err = errors.Wrap(err)
				return
			}
		} else {
			isBoundary = true
		}

		switch {
		case decoder.RequireMetadata && !isBoundary:
			// TODO add context
			err = errors.Wrap(errBoundaryInvalid)
			return

		case !isBoundary:
			state = readerStateSecondBoundary

		default:
			state = readerStateFirstBoundary
			metadataPipe = ohio.MakePipedDecoder(blob, decoder.Metadata)
		}
	}

	var isEOF bool

LINE_READ_LOOP:
	for !isEOF {
		var rawLine, line string

		rawLine, err = bufferedReader.ReadString('\n')
		n += int64(len(rawLine))

		if err == io.EOF {
			err = nil
			isEOF = true
		} else if err != nil {
			err = errors.Wrap(err)
			return
		}

		line = strings.TrimSuffix(rawLine, "\n")

		switch state {
		case readerStateEmpty:
			// nop, processing done above

		case readerStateFirstBoundary:
			if line == Boundary {
				if _, err = metadataPipe.Close(); err != nil {
					err = errors.Wrapf(err, "metadata read failed")
					return
				}

				state += 1
				break
			}

			if _, err = metadataPipe.Write([]byte(rawLine)); err != nil {
				err = errors.Wrap(err)
				return
			}

		case readerStateSecondBoundary:
			break LINE_READ_LOOP

		default:
			err = errors.ErrorWithStackf("impossible state %d", state)
			return
		}
	}

	return
}
