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

func SetHexStringFromAbsolutePath(
	id interfaces.MutableMarklId,
	absOrRelPath string,
	base string,
) (err error) {
	if !filepath.IsAbs(absOrRelPath) {
		return SetHexStringFromRelPath(id, absOrRelPath)
	}

	if absOrRelPath, err = filepath.Rel(base, absOrRelPath); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = SetHexStringFromRelPath(
		id,
		absOrRelPath,
	); err != nil {
		err = errors.Wrapf(
			err,
			"Base: %q",
			absOrRelPath,
			base,
		)

		return
	}

	return
}

func SetHexStringFromRelPath(
	id interfaces.MutableMarklId,
	relPath string,
) (err error) {
	if filepath.IsAbs(relPath) {
		err = errors.Err405MethodNotAllowed.ErrorHiddenf(
			"absolute paths not supported",
		)
		return
	}

	if err = SetHexBytes(
		id.GetMarklFormat().GetMarklFormatId(),
		id,
		[]byte(strings.ReplaceAll(relPath, string(filepath.Separator), "")),
	); err != nil {
		err = errors.Wrapf(
			err,
			"could not transform path into hex. Path: %q",
			relPath,
		)

		return
	}

	return
}

func ReadFrom(reader io.Reader, id *Id, hashType HashType) (n int, err error) {
	id.format = hashType
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
		if id.format, err = GetMarklTypeOrError(tipe); err != nil {
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

		if err = dst.SetMarklId(tipe, bytesDecoded[:numberOfBytesDecoded]); err != nil {
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
		dst.SetMarklId(
			digest.GetMarklFormat().GetMarklFormatId(),
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

	if a.GetMarklFormat() != nil {
		aType = a.GetMarklFormat().GetMarklFormatId()
	}

	if b.GetMarklFormat() != nil {
		bType = b.GetMarklFormat().GetMarklFormatId()
	}

	return aType == bType && bytes.Equal(a.GetBytes(), b.GetBytes())
}

func Clone(src interfaces.MarklId) interfaces.MarklId {
	if !src.IsNull() {
		errors.PanicIfError(MakeErrEmptyType(src))
	}

	if src.GetMarklFormat() == nil {
		panic("empty markl type")
	}

	dst := idPool.Get()
	dst.ResetWithMarklId(src)

	return dst
}

// Creates a human-readable string representation of a digest.
// TODO add type information
func FormatBytesAsHext(merkleId interfaces.MarklId) string {
	return fmt.Sprintf("%x", merkleId.GetBytes())
}

func FormatOrEmptyOnNull(merkleId interfaces.MarklId) string {
	if merkleId.IsNull() {
		return ""
	} else {
		return FormatBytesAsHext(merkleId)
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
