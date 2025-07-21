package sha

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func MakeHashWriter() (h hash.Hash) {
	h = sha256.New()
	return
}

func Make(getter interfaces.DigestGetter) *Sha {
	switch st := getter.GetDigest().(type) {
	case *Sha:
		return st

	default:
		panic(fmt.Sprintf("wrong type: %T", st))
	}
}

func Must(v string) (sh *Sha) {
	sh = poolSha.Get()

	errors.PanicIfError(sh.Set(v))

	return
}

func MakeSha(v string) (sh *Sha, err error) {
	sh = poolSha.Get()

	if err = sh.Set(v); err != nil {
		err = errors.Wrap(err)
	}

	return
}

func MakeShaFromPath(path string) (sh *Sha, err error) {
	sh = poolSha.Get()

	if err = sh.SetFromPath(path); err != nil {
		err = errors.Wrap(err)
	}

	return
}

func FromFormatString(f string, vs ...interface{}) *Sha {
	return FromStringContent(fmt.Sprintf(f, vs...))
}

func FromStringContent(s string) *Sha {
	hash := poolHash256.Get()
	defer poolHash256.Put(hash)

	sr := strings.NewReader(s)

	if _, err := io.Copy(hash, sr); err != nil {
		errors.PanicIfError(err)
	}

	return FromHash(hash)
}

func FromStringer(v interfaces.Stringer) *Sha {
	return FromStringContent(v.String())
}

func FromHash(h hash.Hash) (s *Sha) {
	s = poolSha.Get()
	s.Reset()

	h.Sum(s.data[:0])

	return
}
