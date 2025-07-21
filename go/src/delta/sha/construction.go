package sha

import (
	"fmt"
	"hash"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func MustWithDigest(digest interfaces.Digest) *Sha {
	// TODO instead of checking type, check `GetType()` and use GetBytes
	switch st := digest.(type) {
	case *Sha:
		return st

	default:
		panic(fmt.Sprintf("wrong type: %T", st))
	}
}

func MustWithDigester(digester interfaces.Digester) *Sha {
	return MustWithDigest(digester.GetDigest())
}

func MustWithString(v string) (sh *Sha) {
	sh = poolSha.Get()

	errors.PanicIfError(sh.Set(v))

	return
}

func MakeWithString(v string) (sh *Sha, err error) {
	sh = poolSha.Get()

	if err = sh.Set(v); err != nil {
		err = errors.Wrap(err)
	}

	return
}

func FromFormatString(f string, vs ...any) *Sha {
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
