package merkle

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func CompareToReader(
	reader io.Reader,
	expected interfaces.BlobId,
) int {
	actual := idPool.Get()
	defer idPool.Put(actual)

	actual.data = actual.data[:0]
	actual.data = slices.Grow(actual.data, expected.GetSize())
	actual.data = actual.data[:expected.GetSize()]

	if _, err := io.ReadFull(reader, actual.data); err != nil {
		panic(errors.Wrap(err))
	}

	return bytes.Compare(expected.GetBytes(), actual.GetBytes())
}

func CompareToReaderAt(
	readerAt io.ReaderAt,
	offset int64,
	expected interfaces.BlobId,
) int {
	actual := idPool.Get()
	defer idPool.Put(actual)

	actual.data = actual.data[:0]
	actual.data = slices.Grow(actual.data, expected.GetSize())
	actual.data = actual.data[:expected.GetSize()]

	if _, err := readerAt.ReadAt(actual.data, offset); err != nil {
		panic(errors.Wrap(err))
	}

	return bytes.Compare(expected.GetBytes(), actual.GetBytes())
}

func SetHexBytes(
	tipe string,
	dst interfaces.MutableBlobId,
	bites []byte,
) (err error) {
	bites = bytes.TrimSpace(bites)

	if len(bites) == 0 {
		return
	}

	var numberOfBytesDecoded int
	bytesDecoded := make([]byte, hex.DecodedLen(len(bites)))

	if numberOfBytesDecoded, err = hex.Decode(bytesDecoded, bites); err != nil {
		err = errors.Wrapf(err, "N: %d, Data: %q", numberOfBytesDecoded, bites)
		return
	}

	if err = dst.SetMerkleId(tipe, bytesDecoded[:numberOfBytesDecoded]); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func SetDigester(
	dst interfaces.MutableBlobId,
	src interfaces.BlobIdGetter,
) {
	digest := src.GetBlobId()
	errors.PanicIfError(dst.SetMerkleId(digest.GetType(), digest.GetBytes()))
}

func EqualsReader(
	expectedBlobId interfaces.BlobId,
	bufferedReader *bufio.Reader,
) (ok bool, err error) {
	var actualBytes []byte
	bytesRead := expectedBlobId.GetSize()

	if actualBytes, err = bufferedReader.Peek(bytesRead); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	ok = bytes.Equal(expectedBlobId.GetBytes(), actualBytes)

	if _, err = bufferedReader.Discard(bytesRead); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}

func Equals(a, b interfaces.BlobId) bool {
	return (a.IsNull() && b.IsNull()) ||
		(a.GetType() == b.GetType() && bytes.Equal(a.GetBytes(), b.GetBytes()))
}

func Clone(src interfaces.BlobId) interfaces.BlobId {
	if src.IsNull() {
		return Null
	}

	errors.PanicIfError(MakeErrEmptyType(src))
	env := GetEnv(src.GetType())
	dst := env.GetBlobId()
	dst.ResetWithMerkleId(src)

	return dst
}

// Creates a human-readable string representation of a digest.
// TODO add type information
func Format(merkleId interfaces.BlobId) string {
	return fmt.Sprintf("%x", merkleId.GetBytes())
}

func FormatOrEmptyOnNull(merkleId interfaces.BlobId) string {
	if merkleId.IsNull() {
		return ""
	} else {
		return Format(merkleId)
	}
}
