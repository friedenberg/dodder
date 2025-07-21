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
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
)

const (
	ByteSize      = 32
	ShaNullString = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	Null          = ShaNullString
)

type Bytes [ByteSize]byte

var shaNull Sha

func init() {
	errors.PanicIfError(shaNull.Set(ShaNullString))
}

type PathComponents interface {
	PathComponents() []string
}

type ShaLike = interfaces.ShaGetter

type Sha struct {
	data *Bytes
}

func (s *Sha) Size() int {
	return ByteSize
}

func (s *Sha) GetBytes() []byte {
	return s.GetShaBytes()
}

func (s *Sha) GetShaBytes() []byte {
	if s.IsNull() {
		return shaNull.data[:]
	} else {
		return s.data[:]
	}
}

func (s *Sha) GetShaString() string {
	if s == nil || s.IsNull() {
		return fmt.Sprintf("%x", shaNull.data[:])
	} else {
		return fmt.Sprintf("%x", s.data[:])
	}
}

func (s *Sha) String() string {
	return s.GetShaString()
}

func (s *Sha) Sha() *Sha {
	return s
}

func (src *Sha) WriteTo(w io.Writer) (n int64, err error) {
	var n1 int
	n1, err = ohio.WriteAllOrDieTrying(w, src.GetShaBytes())
	n = int64(n1)
	return
}

func (s *Sha) GetShaLike() interfaces.Sha {
	return s
}

func (s *Sha) IsNull() bool {
	if s == nil || s.data == nil {
		return true
	}

	if bytes.Equal(s.data[:], shaNull.data[:]) {
		return true
	}

	return false
}

func (s *Sha) GetHead() string {
	return s.String()[0:2]
}

func (s *Sha) GetTail() string {
	return s.String()[2:]
}

func (sha *Sha) AssertEqualsShaLike(b interfaces.Sha) error {
	if !sha.EqualsSha(b) {
		return MakeErrNotEqual(sha, b)
	}

	return nil
}

func (sha *Sha) EqualsAny(b any) bool {
	return values.Equals(sha, b)
}

func (sha *Sha) EqualsSha(b interfaces.Sha) bool {
	return sha.GetShaString() == b.GetShaString()
}

func (sha *Sha) Equals(b *Sha) bool {
	return sha.GetShaString() == b.GetShaString()
}

//  __        __    _ _   _
//  \ \      / / __(_) |_(_)_ __   __ _
//   \ \ /\ / / '__| | __| | '_ \ / _` |
//    \ V  V /| |  | | |_| | | | | (_| |
//     \_/\_/ |_|  |_|\__|_|_| |_|\__, |
//                                |___/

func (dst *Sha) SetFromHash(h hash.Hash) (err error) {
	dst.allocDataIfNecessary()
	b := h.Sum(dst.data[:0])
	err = makeErrLength(ByteSize, len(b))
	return
}

func (dst *Sha) SetShaLike(src ShaLike) (err error) {
	dst.allocDataIfNecessary()

	err = makeErrLength(
		ByteSize,
		copy(dst.data[:], src.GetShaLike().GetShaBytes()),
	)

	// if dst.String() ==
	// "f32cf7f2b1b8b7688c78dec4eb3c3675fcdc3aeaaf4c1305f3bdaf7fc0252e02" {
	// 	panic("found")
	// }

	return
}

func (s *Sha) SetParts(head, tail string) (err error) {
	if err = s.Set(head + tail); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Sha) SetFromPath(path string) (err error) {
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

	if err = s.SetParts(head, tail); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Sha) ReadAtFrom(
	reader io.ReaderAt,
	start int64,
) (bytesRead int64, err error) {
	s.allocDataIfNecessary()

	var bytesCount int
	bytesCount, err = reader.ReadAt(s.data[:], start)
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

func (s *Sha) ReadFrom(reader io.Reader) (bytesRead int64, err error) {
	s.allocDataIfNecessary()

	var bytesCount int
	bytesCount, err = ohio.ReadAllOrDieTrying(reader, s.data[:])
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

func (s *Sha) SetHexBytes(bytess []byte) (err error) {
	s.allocDataIfNecessary()

	bytess = bytes.TrimSpace(bytess)

	if len(bytess) == 0 {
		return
	}

	var bytesDecoded int

	if _, bytesDecoded, err = hexDecode(s.data[:0], bytess); err != nil {
		err = errors.Wrapf(err, "N: %d, Data: %q", bytesDecoded, bytess)
		return
	}

	return
}

func (s *Sha) Set(value string) (err error) {
	s.allocDataIfNecessary()

	value1 := strings.TrimSpace(value)
	value1 = strings.TrimPrefix(value1, "@")

	var decodedBytes []byte

	if decodedBytes, err = hex.DecodeString(value1); err != nil {
		err = errors.Wrapf(err, "%q", value1)
		return
	}

	bytesWritten := copy(s.data[:], decodedBytes)

	if err = makeErrLength(ByteSize, bytesWritten); err != nil {
		return
	}

	return
}

func (s *Sha) allocDataIfNecessary() {
	if s.data != nil {
		return
	}

	s.data = &Bytes{}
}

func (s *Sha) Reset() {
	s.allocDataIfNecessary()
	s.ResetWith(&shaNull)
}

func (sha *Sha) ResetWith(other *Sha) {
	sha.allocDataIfNecessary()

	if other.IsNull() {
		copy(sha.data[:], shaNull.data[:])
	} else {
		copy(sha.data[:], other.data[:])
	}
}

func (sha *Sha) ResetWithShaLike(other interfaces.Sha) {
	sha.allocDataIfNecessary()
	copy(sha.data[:], other.GetShaBytes())
}

func (s *Sha) Path(pathComponents ...string) string {
	pathComponents = append(pathComponents, s.GetHead())
	pathComponents = append(pathComponents, s.GetTail())

	return path.Join(pathComponents...)
}

func (s *Sha) MarshalBinary() (text []byte, err error) {
	text = []byte(s.String())

	return
}

func (s *Sha) UnmarshalBinary(text []byte) (err error) {
	if err = s.Set(string(text)); err != nil {
		return
	}

	return
}

func (s *Sha) MarshalText() (text []byte, err error) {
	text = []byte(s.String())

	return
}

func (s *Sha) UnmarshalText(text []byte) (err error) {
	if err = s.Set(string(text)); err != nil {
		return
	}

	return
}
