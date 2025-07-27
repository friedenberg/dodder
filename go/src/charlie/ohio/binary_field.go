package ohio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type BinaryField struct {
	ContentLength [2]uint8
	Content       bytes.Buffer
}

func (bf *BinaryField) String() string {
	cl, _, _ := bf.GetContentLength()
	return fmt.Sprintf("%d:%x", cl, bf.Content.Bytes())
}

func (bf *BinaryField) Reset() {
	bf.ContentLength[0] = 0
	bf.ContentLength[1] = 0
	bf.Content.Reset()
}

func (bf *BinaryField) GetContentLength() (contentLength int, contentLength64 int64, err error) {
	var n int
	contentLength64, n = binary.Varint(bf.ContentLength[:])

	if n <= 0 {
		err = errors.ErrorWithStackf("error in content length: %d", n)
		return
	}

	if contentLength64 > math.MaxUint16 {
		err = errContentLengthTooLarge
		return
	}

	if contentLength64 < 0 {
		err = errContentLengthNegative
		return
	}

	return int(contentLength64), contentLength64, nil
}

func (bf *BinaryField) SetContentLength(v int) {
	if v < 0 {
		panic(errContentLengthNegative)
	}

	if v > math.MaxUint16 {
		panic(errContentLengthTooLarge)
	}

	// TODO
	binary.PutVarint(bf.ContentLength[:], int64(v))
}

var (
	errContentLengthTooLarge = errors.New("content length too large")
	errContentLengthNegative = errors.New("content length negative")
)

func (bf *BinaryField) ReadFrom(r io.Reader) (n int64, err error) {
	var n1 int
	var n2 int64

	n1, err = ReadAllOrDieTrying(r, bf.ContentLength[:])
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	contentLength, contentLength64, err := bf.GetContentLength()
	if err != nil {
		err = errors.Wrap(err)
		return
	}

	bf.Content.Grow(contentLength)
	bf.Content.Reset()

	n2, err = io.CopyN(&bf.Content, r, contentLength64)
	n += n2

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}

var errContentLengthDoesNotMatchContent = errors.New(
	"content length does not match content",
)

func (bf *BinaryField) WriteTo(w io.Writer) (n int64, err error) {
	if bf.Content.Len() > math.MaxUint16 {
		err = errContentLengthTooLarge
		return
	}

	bf.SetContentLength(bf.Content.Len())

	var n1 int
	var n2 int64

	n1, err = WriteAllOrDieTrying(w, bf.ContentLength[:])
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	n2, err = io.Copy(w, &bf.Content)
	n += n2

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}
