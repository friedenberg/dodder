package triple_hyphen_io

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
)

const (
	Boundary = "---"
)

var BoundaryStringValue catgut.String

func init() {
	errors.PanicIfError(BoundaryStringValue.Set(Boundary))
}

var errBoundaryInvalid = errors.New("boundary invalid")

type Peeker interface {
	io.Reader
	Peek(n int) (bytes []byte, err error)
}

func ReadBoundaryFromPeeker(peeker Peeker) (err error) {
	if err = PeekBoundaryFromPeeker(peeker); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF, errBoundaryInvalid)
		return
	}

	if _, err = io.CopyN(
		io.Discard,
		peeker,
		int64(BoundaryStringValue.Len()+1),
	); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}

func PeekBoundaryFromPeeker(peeker Peeker) (err error) {
	var peeked []byte

	if peeked, err = peeker.Peek(BoundaryStringValue.Len() + 1); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	boundaryOnly := peeked[:BoundaryStringValue.Len()]
	newline := peeked[BoundaryStringValue.Len()]

	if !BoundaryStringValue.EqualsBytes(boundaryOnly) {
		err = errBoundaryInvalid
		return
	}

	if newline != '\n' {
		err = errBoundaryInvalid
		return
	}

	return
}

func ReadBoundaryFromRingBuffer(
	ringBuffer *catgut.RingBuffer,
) (err error) {
	var readable catgut.Slice

	readable, err = ringBuffer.PeekUpto('\n')

	if errors.IsNotNilAndNotEOF(err) {
		return
	} else if readable.Len() == 0 && err == io.EOF {
		return
	}

	if readable.Len() != BoundaryStringValue.Len() {
		err = errors.Wrapf(
			errBoundaryInvalid,
			"Expected: %d, Actual: %d. Readable: %q",
			BoundaryStringValue.Len(),
			readable.Len(),
			readable,
		)

		return
	}

	if !readable.Equal(BoundaryStringValue.Bytes()) {
		err = errors.Wrapf(
			errBoundaryInvalid,
			"Expected: %q, Actual: %q",
			BoundaryStringValue.String(),
			readable.String(),
		)
		return
	}

	// boundary and newline
	n := BoundaryStringValue.Len() + 1
	ringBuffer.AdvanceRead(n)

	if err == io.EOF {
		err = nil
	}

	return
}
