package markl

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
)

func SetHexStringFromPath(
	id interfaces.MutableMarklId,
	path string,
) (err error) {
	tail := filepath.Base(path)
	head := filepath.Base(filepath.Dir(path))

	switch {
	case tail == string(filepath.Separator) || head == string(filepath.Separator):
		fallthrough

	case tail == "." || head == ".":
		err = errors.ErrorWithStackf(
			"path cannot be turned into a head/tail pair: '%s/%s'",
			head,
			tail,
		)

		return
	}

	if err = SetHexBytes(id.GetMarklType().GetMarklTypeId(), id, []byte(head+tail)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func ReadFrom(reader io.Reader, id *Id, hashType HashType) (n int, err error) {
	id.tipe = hashType
	id.allocDataAndSetToCapIfNecessary(hashType.GetSize())

	if n, err = io.ReadFull(reader, id.data); err != nil {
		errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}

func CompareToReader(
	reader io.Reader,
	expected interfaces.MarklId,
) int {
	actual := idPool.Get()
	defer idPool.Put(actual)

	actual.allocDataAndSetToCapIfNecessary(expected.GetSize())

	if _, err := io.ReadFull(reader, actual.data); err != nil {
		panic(errors.Wrap(err))
	}

	return bytes.Compare(expected.GetBytes(), actual.GetBytes())
}

func CompareToReaderAt(
	readerAt io.ReaderAt,
	offset int64,
	expected interfaces.MarklId,
) int {
	actual := idPool.Get()
	defer idPool.Put(actual)

	actual.allocDataAndSetToCapIfNecessary(expected.GetSize())

	if _, err := readerAt.ReadAt(actual.data, offset); err != nil {
		panic(errors.Wrap(err))
	}

	return bytes.Compare(expected.GetBytes(), actual.GetBytes())
}

func SetHexBytes(
	tipe string,
	dst interfaces.MutableMarklId,
	bites []byte,
) (err error) {
	bites = bytes.TrimSpace(bites)

	if id, ok := dst.(*Id); ok {
		if id.tipe, err = GetMarklTypeOrError(tipe); err != nil {
			err = errors.Wrapf(
				err,
				"failed to SetHexBytes with type %s and bites %s",
				tipe,
				bites,
			)
			return
		}

		id.allocDataAndSetToCapIfNecessary(hex.DecodedLen(len(bites)))

		var numberOfBytesDecoded int

		numberOfBytesDecoded, err = hex.Decode(id.data, bites)
		id.data = id.data[:numberOfBytesDecoded]

		if err != nil {
			err = errors.Wrapf(
				err,
				"N: %d, Data: %q",
				numberOfBytesDecoded,
				bites,
			)
			return
		}
	} else {
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
	}

	return
}

func SetDigester(
	dst interfaces.MutableMarklId,
	src interfaces.MarklIdGetter,
) {
	digest := src.GetMarklId()
	errors.PanicIfError(
		dst.SetMerkleId(
			digest.GetMarklType().GetMarklTypeId(),
			digest.GetBytes(),
		),
	)
}

func EqualsReader(
	expectedBlobId interfaces.MarklId,
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

func Equals(a, b interfaces.MarklId) bool {
	if a.IsNull() && b.IsNull() {
		return true
	}

	var aType, bType string

	if a.GetMarklType() != nil {
		aType = a.GetMarklType().GetMarklTypeId()
	}

	if b.GetMarklType() != nil {
		bType = b.GetMarklType().GetMarklTypeId()
	}

	return aType == bType && bytes.Equal(a.GetBytes(), b.GetBytes())
}

func Clone(src interfaces.MarklId) interfaces.MarklId {
	if !src.IsNull() {
		errors.PanicIfError(MakeErrEmptyType(src))
	}

	if src.GetMarklType() == nil {
		panic("empty markl type")
	}

	dst := idPool.Get()
	dst.ResetWithMarklId(src)

	return dst
}

// Creates a human-readable string representation of a digest.
// TODO add type information
func Format(merkleId interfaces.MarklId) string {
	return fmt.Sprintf("%x", merkleId.GetBytes())
}

func FormatOrEmptyOnNull(merkleId interfaces.MarklId) string {
	if merkleId.IsNull() {
		return ""
	} else {
		return Format(merkleId)
	}
}

func SetFromPath(id interfaces.MutableMarklId, path string) (err error) {
	var file *os.File

	if file, err = files.Open(path); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, file)

	bufferedReader, repool := pool.GetBufferedReader(file)
	defer repool()

	var isEOF bool
	var key string

	for !isEOF {
		var line string
		line, err = bufferedReader.ReadString('\n')

		if err == io.EOF {
			isEOF = true
			err = nil
		} else if err != nil {
			err = errors.Wrap(err)
			return
		}

		if len(line) > 0 {
			key = strings.TrimSpace(line)
		}
	}

	// if !strings.HasPrefix(maybeBech32String, "AGE-SECRET-KEY-1") {
	// 	value = maybeBech32String
	// }

	// var data []byte

	// if _, data, err = bech32.Decode(maybeBech32String); err != nil {
	// 	err = errors.Wrap(err)
	// 	return
	// }

	// if err = blobStoreConfig.Encryption.SetFormat(
	// 	markl.FormatIdMadderPrivateKeyV0,
	// ); err != nil {
	// 	err = errors.Wrap(err)
	// 	return
	// }

	// if err = blobStoreConfig.Encryption.SetMerkleId(
	// 	markl.TypeIdAgeX25519Sec,
	// 	data,
	// ); err != nil {
	// 	err = errors.Wrap(err)
	// 	return
	// }

	if err = id.Set(key); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
