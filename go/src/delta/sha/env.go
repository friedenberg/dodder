package sha

import (
	"fmt"
	"hash"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type Env struct{}

func (env Env) MakeWriteDigester() interfaces.WriteDigester {
	return MakeWriter(nil)
}

func (env Env) MakeReadDigester() interfaces.ReadDigester {
	return MakeReadCloser(nil)
}

// TODO switch to being functions on Env that return interfaces.Digest

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

func FromHash(hash hash.Hash) (digest *Sha) {
	digest = poolSha.Get()
	digest.Reset()

	if hash.Size() != ByteSize {
		panic(
			fmt.Sprintf(
				"expected hash size to be %d but was %d. Hash: %T",
				ByteSize,
				hash.Size(),
				hash,
			),
		)
	}

	// the return value isn't used because s.data is already the right size
	hash.Sum(digest.data[:0])

	return
}
