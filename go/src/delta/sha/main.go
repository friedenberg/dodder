package sha

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"path"
	"path/filepath"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/hecks"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
)

const (
	Type       = "sha256"
	ByteSize   = 32
	nullString = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

type byteArray [ByteSize]byte

var null Sha

func init() {
	errors.PanicIfError(null.Set(nullString))

	if !null.IsNull() {
		panic("null digest is not null")
	}
}

type PathComponents interface {
	PathComponents() []string
}

// TODO rename to digest
type Sha struct {
	nonZero bool
	data    byteArray
}

func (digest *Sha) GetSize() int {
	return ByteSize
}

func (digest *Sha) GetBytes() []byte {
	if digest.IsNull() {
		return null.data[:]
	} else {
		return digest.data[:]
	}
}

func (digest *Sha) String() string {
	if digest == nil || digest.IsNull() {
		return fmt.Sprintf("%x", null.data[:])
	} else {
		return fmt.Sprintf("%x", digest.data[:])
	}
}

func (digest *Sha) WriteTo(w io.Writer) (n int64, err error) {
	var n1 int
	n1, err = ohio.WriteAllOrDieTrying(w, digest.GetBytes())
	n = int64(n1)
	return
}

func (digest *Sha) GetType() string {
	return Type
}

func (digest *Sha) GetBlobId() interfaces.MerkleId {
	return digest
}

func (digest *Sha) IsNull() bool {
	if digest == nil {
		return true
	}

	if !digest.nonZero {
		return true
	}

	if bytes.Equal(digest.data[:], null.data[:]) {
		return true
	}

	return false
}

func (digest *Sha) GetHead() string {
	return digest.String()[0:2]
}

func (digest *Sha) GetTail() string {
	return digest.String()[2:]
}

// func (digest *Sha) EqualsAny(b any) bool {
// 	return values.Equals(digest, b)
// }

// func (digest *Sha) Equals(b *Sha) bool {
// 	return digest.String() == b.String()
// }

//  __        __    _ _   _
//  \ \      / / __(_) |_(_)_ __   __ _
//   \ \ /\ / / '__| | __| | '_ \ / _` |
//    \ V  V /| |  | | |_| | | | | (_| |
//     \_/\_/ |_|  |_|\__|_|_| |_|\__, |
//                                |___/

func (digest *Sha) SetFromHash(h hash.Hash) (err error) {
	digest.setNonZero()
	b := h.Sum(digest.data[:0])
	err = merkle_ids.MakeErrLength(ByteSize, len(b))
	return
}

func (digest *Sha) SetDigester(src interfaces.BlobIdGetter) (err error) {
	return digest.SetDigest(src.GetBlobId())
}

// TODO replace
func (digest *Sha) SetDigest(src interfaces.BlobId) (err error) {
	return digest.SetMerkleId(src.GetType(), src.GetBytes())
}

func (digest *Sha) SetMerkleId(tipe string, bites []byte) (err error) {
	if tipe != digest.GetType() {
		err = errors.Errorf(
			"cannot set digest from type %q, need %q",
			tipe,
			digest.GetType(),
		)

		return
	}

	digest.SetBytes(bites)

	return
}

func (digest *Sha) SetBytes(bytess []byte) (err error) {
	digest.setNonZero()

	err = merkle_ids.MakeErrLength(
		ByteSize,
		copy(digest.data[:], bytess),
	)

	return
}

// TODO remove
func (digest *Sha) SetParts(head, tail string) (err error) {
	if err = digest.Set(head + tail); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO remove
func (digest *Sha) SetFromPath(path string) (err error) {
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

	if err = digest.SetParts(head, tail); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (digest *Sha) ReadAtFrom(
	reader io.ReaderAt,
	start int64,
) (bytesRead int64, err error) {
	digest.setNonZero()

	var bytesCount int
	bytesCount, err = reader.ReadAt(digest.data[:], start)
	bytesRead += int64(bytesCount)

	if bytesRead == 0 && err == io.EOF {
		return
	} else if bytesRead != ByteSize && bytesRead != 0 {
		err = errors.ErrorWithStackf("expected to read %d bytes but only read %d", ByteSize, bytesRead)
		return
	} else if errors.IsNotNilAndNotEOF(err) {
		err = errors.Wrap(err)
		return
	}

	return
}

func (digest *Sha) ReadFrom(reader io.Reader) (bytesRead int64, err error) {
	digest.setNonZero()

	var bytesCount int
	bytesCount, err = ohio.ReadAllOrDieTrying(reader, digest.data[:])
	bytesRead += int64(bytesCount)

	if bytesRead == 0 && err == io.EOF {
		return
	} else if bytesRead != ByteSize && bytesRead != 0 {
		err = errors.ErrorWithStackf("expected to read %d bytes but only read %d", ByteSize, bytesRead)
		return
	} else if errors.IsNotNilAndNotEOF(err) {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO move to digests package
func (digest *Sha) SetHexBytes(bytess []byte) (err error) {
	digest.setNonZero()

	bytess = bytes.TrimSpace(bytess)

	if len(bytess) == 0 {
		return
	}

	var bytesDecoded int

	if bytesDecoded, err = hecks.Decode(digest.data[:0], bytess); err != nil {
		err = errors.Wrapf(err, "N: %d, Data: %q", bytesDecoded, bytess)
		return
	}

	return
}

func (digest *Sha) Set(value string) (err error) {
	digest.setNonZero()

	value1 := strings.TrimSpace(value)
	value1 = strings.TrimPrefix(value1, "@")

	var decodedBytes []byte

	if decodedBytes, err = hex.DecodeString(value1); err != nil {
		err = errors.Wrapf(err, "%q", value1)
		return
	}

	bytesWritten := copy(digest.data[:], decodedBytes)

	if err = merkle_ids.MakeErrLength(ByteSize, bytesWritten); err != nil {
		return
	}

	return
}

func (digest *Sha) setNonZero() {
	digest.nonZero = true
}

func (digest *Sha) Reset() {
	digest.setNonZero()
	digest.ResetWith(&null)
}

func (digest *Sha) ResetWith(other *Sha) {
	digest.setNonZero()

	if other.IsNull() {
		copy(digest.data[:], null.data[:])
	} else {
		copy(digest.data[:], other.data[:])
	}
}

func (digest *Sha) ResetWithMerkleId(src interfaces.MerkleId) {
	errors.PanicIfError(digest.SetMerkleId(src.GetType(), src.GetBytes()))
}

func (digest *Sha) ResetWithShaLike(other interfaces.BlobId) {
	digest.setNonZero()
	copy(digest.data[:], other.GetBytes())
}

func (digest *Sha) Path(pathComponents ...string) string {
	pathComponents = append(pathComponents, digest.GetHead())
	pathComponents = append(pathComponents, digest.GetTail())

	return path.Join(pathComponents...)
}

func (digest *Sha) MarshalBinary() (text []byte, err error) {
	text = []byte(digest.String())

	return
}

func (digest *Sha) UnmarshalBinary(text []byte) (err error) {
	if err = digest.Set(string(text)); err != nil {
		return
	}

	return
}

func (digest *Sha) MarshalText() (text []byte, err error) {
	text = []byte(digest.String())

	return
}

func (digest *Sha) UnmarshalText(text []byte) (err error) {
	if err = digest.Set(string(text)); err != nil {
		return
	}

	return
}
