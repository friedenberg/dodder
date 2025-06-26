package ohio

import (
	"encoding/binary"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

func ReadAllOrDieTrying(r io.Reader, b []byte) (n int, err error) {
	var acc int

	for n < len(b) {
		acc, err = r.Read(b[n:])
		n += acc

		if err != nil {
			break
		}
	}

	switch err {
	case io.EOF:
		if n < len(b) && n > 0 {
			err = errors.Wrapf(
				io.ErrUnexpectedEOF,
				"Expected %d, got %d",
				len(b),
				n,
			)
		}
	case nil:
	default:
		err = errors.Wrap(err)
	}

	return
}

func ReadUint8(r io.Reader) (n uint8, read int, err error) {
	cl := [1]byte{}

	read, err = ReadAllOrDieTrying(r, cl[:])
	if err != nil {
		err = errors.WrapExcept(err, io.EOF)
		return
	}

	clInt, clIntErr := binary.Uvarint(cl[:])

	if clIntErr <= 0 {
		err = errors.WrapExcept(err, io.EOF)
		return
	}

	n = uint8(clInt)

	return
}

func ReadFixedUint8(r io.Reader) (n uint8, read int, err error) {
	cl := [1]byte{}

	read, err = ReadAllOrDieTrying(r, cl[:])
	if err != nil {
		err = errors.WrapExcept(err, io.EOF)
		return
	}

	n = cl[0]

	return
}

func ReadInt8(r io.Reader) (n int8, read int, err error) {
	cl := [1]byte{}

	read, err = ReadAllOrDieTrying(r, cl[:])
	if err != nil {
		err = errors.WrapExcept(err, io.EOF)
		return
	}

	clInt, clIntErr := binary.Uvarint(cl[:])

	if clIntErr <= 0 {
		err = errors.WrapExcept(err, io.EOF)
		return
	}

	n = int8(clInt)

	return
}

func ReadUint16(r io.ByteReader) (v uint16, n int64, err error) {
	var clInt uint64

	if clInt, err = binary.ReadUvarint(r); err != nil {
		err = errors.Wrap(err)
		return
	}

	v = uint16(clInt)
	n = int64(binary.Size(v))

	return
}

func ReadInt64(r io.Reader) (n int64, read int, err error) {
	cl := [binary.MaxVarintLen64]byte{}

	read, err = ReadAllOrDieTrying(r, cl[:])
	if err != nil {
		err = errors.WrapExcept(err, io.EOF)
		return
	}

	n, clIntErr := binary.Varint(cl[:])

	if clIntErr <= 0 {
		err = errors.WrapExcept(err, io.EOF)
		return
	}

	return
}

func ReadFixedUInt16(r io.Reader) (n int, val uint16, err error) {
	cl := [2]byte{}

	n, err = ReadAllOrDieTrying(r, cl[:])
	if err != nil {
		err = errors.WrapExcept(err, io.EOF)
		return
	}

	val = ByteArrayToUInt16(cl)

	return
}

func ReadFixedInt32(r io.Reader) (n int, val int32, err error) {
	cl := [4]byte{}

	n, err = ReadAllOrDieTrying(r, cl[:])
	if err != nil {
		err = errors.WrapExcept(err, io.EOF)
		return
	}

	val = ByteArrayToInt32(cl)

	return
}

func ReadFixedInt64(r io.Reader) (n int, val int64, err error) {
	cl := [8]byte{}

	n, err = ReadAllOrDieTrying(r, cl[:])
	if err != nil {
		err = errors.WrapExcept(err, io.EOF)
		return
	}

	val = ByteArrayToInt64(cl)

	return
}
