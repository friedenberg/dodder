package merkle_ids

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/hecks"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func SetHex(
	tipe string,
	dst interfaces.MutableMerkleId,
	hex string,
) (err error) {
	hex = strings.TrimSpace(hex)

	if len(hex) == 0 {
		return
	}

	var bytesDecoded int

	if bytesDecoded, err = hecks.Decode(dst.GetBytes(), []byte(hex)); err != nil {
		err = errors.Wrapf(err, "N: %d, Data: %q", bytesDecoded, hex)
		return
	}

	return
}

func SetHexBytes(
	tipe string,
	dst interfaces.MutableMerkleId,
	bites []byte,
) (err error) {
	bites = bytes.TrimSpace(bites)

	if len(bites) == 0 {
		return
	}

	var bytesDecoded int

	if bytesDecoded, err = hecks.Decode(dst.GetBytes(), bites); err != nil {
		err = errors.Wrapf(err, "N: %d, Data: %q", bytesDecoded, bites)
		return
	}

	return
}

func SetDigester(dst interfaces.MutableMerkleId, src interfaces.BlobIdGetter) {
	blobId := src.GetBlobId()
	dst.SetMerkleId(blobId.GetType(), blobId.GetBytes())
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

func Equals(a, b interfaces.MerkleId) bool {
	return a.GetType() == b.GetType() && bytes.Equal(a.GetBytes(), b.GetBytes())
}

func Clone(src interfaces.BlobId) interfaces.BlobId {
	env := GetEnv(src.GetType())
	dst := env.GetBlobId()
	errors.PanicIfError(dst.SetDigest(src))
	return dst
}

// Creates a human-readable string representation of a digest.
// TODO add type information
func Format(merkleId interfaces.MerkleId) string {
	return fmt.Sprintf("%x", merkleId.GetBytes())
}
