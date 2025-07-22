package sha

import (
	"fmt"
	"hash"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
)

type Env struct{}

func (env Env) GetHash() hash.Hash {
	return poolHash256.Get()
}

func (env Env) PutHash(hash hash.Hash) {
	poolHash256.Put(hash)
}

func (env Env) GetDigest() interfaces.Digest {
	return poolSha.Get()
}

func (env Env) PutDigest(digest interfaces.Digest) {
	poolSha.Put(digest.(*Sha))
}

func (env Env) MakeDigestFromHash(hash hash.Hash) (interfaces.Digest, error) {
	digest := poolSha.Get()
	digest.Reset()

	if err := digests.MakeErrLength(ByteSize, hash.Size()); err != nil {
		return nil, err
	}

	// the return value isn't used because s.data is already the right size
	hash.Sum(digest.data[:0])

	return digest, nil
}

func (env Env) MakeWriteDigester() interfaces.WriteDigester {
	return MakeWriter(env, nil)
}

// TODO switch to being functions on Env that return interfaces.Digest

func FromFormatString(f string, vs ...any) interfaces.Digest {
	return FromStringContent(fmt.Sprintf(f, vs...))
}

func FromStringContent(s string) interfaces.Digest {
	hash := poolHash256.Get()
	defer poolHash256.Put(hash)

	stringReader := strings.NewReader(s)

	if _, err := io.Copy(hash, stringReader); err != nil {
		errors.PanicIfError(err)
	}

	return FromHash(hash)
}

func FromStringer(v interfaces.Stringer) interfaces.Digest {
	return FromStringContent(v.String())
}

func FromHash(hash hash.Hash) interfaces.Digest {
	digest, err := Env{}.MakeDigestFromHash(hash)
	if err != nil {
		panic(err)
	}

	return digest
}
