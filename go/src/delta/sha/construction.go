package sha

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"path/filepath"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func MakeHashWriter() (h hash.Hash) {
	h = sha256.New()
	return
}

func Make(getter interfaces.ShaGetter) *Sha {
	switch st := getter.GetShaLike().(type) {
	case *Sha:
		return st

	default:
		panic(fmt.Sprintf("wrong type: %T", st))
	}
}

func Must(v string) (s *Sha) {
	s = poolSha.Get()

	errors.PanicIfError(s.Set(v))

	return
}

func MakeSha(v string) (s *Sha, err error) {
	s = poolSha.Get()

	if err = s.Set(v); err != nil {
		err = errors.Wrap(err)
	}

	return
}

func MakeShaFromPath(p string) (s *Sha, err error) {
	tail := filepath.Base(p)
	head := filepath.Base(filepath.Dir(p))

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

	if s, err = MakeSha(fmt.Sprintf("%s%s", head, tail)); err != nil {
		err = errors.Wrapf(err, "head: %q, tail: %q", head, tail)
		return
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
