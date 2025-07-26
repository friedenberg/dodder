package sha

import (
	"fmt"
	"hash"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var _ = digests.RegisterEnv(Env{})

type Env struct{}

func (env Env) GetType() string {
	return Type
}

func (env Env) GetHash() (hash.Hash, func()) {
	return pool.GetSha256Hash()
}

func (env Env) GetBlobId() interfaces.BlobId {
	return poolSha.Get()
}

func (env Env) PutBlobId(digest interfaces.BlobId) {
	poolSha.Put(digest.(*Sha))
}

func (env Env) MakeDigestFromString(
	value string,
) (interfaces.BlobId, interfaces.FuncRepool, error) {
	digest := poolSha.Get()
	digest.Reset()

	if err := digest.Set(value); err != nil {
		poolSha.Put(digest)
		return nil, nil, err
	}

	return digest, func() { poolSha.Put(digest) }, nil
}

func (env Env) MakeDigestFromHash(hash hash.Hash) (interfaces.BlobId, error) {
	digest := poolSha.Get()
	digest.Reset()

	if err := digests.MakeErrLength(ByteSize, hash.Size()); err != nil {
		return nil, err
	}

	// the return value isn't used because s.data is already the right size
	hash.Sum(digest.data[:0])

	return digest, nil
}

func (env Env) MakeWriteDigesterWithRepool() (interfaces.WriteDigester, func()) {
	return digests.MakeWriterWithRepool(env, nil)
}

func (env Env) MakeWriteDigester() interfaces.WriteDigester {
	return digests.MakeWriter(env, nil)
}

// TODO switch to being functions on Env that return interfaces.Digest

func FromFormatString(f string, vs ...any) interfaces.BlobId {
	return FromStringContent(fmt.Sprintf(f, vs...))
}

func FromStringContent(s string) interfaces.BlobId {
	hash, repool := pool.GetSha256Hash()
	defer repool()

	stringReader, repool2 := pool.GetStringReader(s)
	defer repool2()

	if _, err := io.Copy(hash, stringReader); err != nil {
		errors.PanicIfError(err)
	}

	return FromHash(hash)
}

func FromStringer(v interfaces.Stringer) interfaces.BlobId {
	return FromStringContent(v.String())
}

func FromHash(hash hash.Hash) interfaces.BlobId {
	digest, err := Env{}.MakeDigestFromHash(hash)
	if err != nil {
		panic(err)
	}

	return digest
}
