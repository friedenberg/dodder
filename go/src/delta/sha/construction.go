package sha

import (
	"fmt"

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
