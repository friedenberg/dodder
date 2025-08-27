package sha

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/pool_value"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
)

var (
	Env = env{tipe: Type}
	_   = merkle_ids.RegisterEnv(Env)
	_   = merkle_ids.RegisterEnv(env{merkle.HRPObjectBlobDigestSha256V1})
	_   = merkle_ids.RegisterEnv(env{merkle.HRPObjectDigestSha256V1})

	sha256Hash = pool_value.Make(
		func() hash.Hash {
			return sha256.New()
		},
		func(hash hash.Hash) {
			hash.Reset()
		},
	)

	poolSha = pool.Make(
		nil,
		func(sh *Sha) {
			sh.Reset()
		},
	)
)

type env struct {
	tipe string
}

func (env env) GetType() string {
	return env.tipe
}

func (env env) GetHash() (hash.Hash, interfaces.FuncRepool) {
	return pool.GetSha256Hash()
}

func (env env) GetBlobId() interfaces.MutableBlobId {
	return poolSha.Get()
}

func (env env) PutBlobId(digest interfaces.BlobId) {
	poolSha.Put(digest.(*Sha))
}

func (env env) MakeDigestFromString(
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

func (env env) MakeDigestFromHash(hash hash.Hash) (interfaces.BlobId, error) {
	digest := poolSha.Get()
	digest.Reset()

	if err := merkle_ids.MakeErrLength(ByteSize, hash.Size()); err != nil {
		return nil, err
	}

	// the return value isn't used because s.data is already the right size
	hash.Sum(digest.data[:0])

	return digest, nil
}

func (env env) MakeWriteDigesterWithRepool() (interfaces.WriteBlobIdGetter, interfaces.FuncRepool) {
	return merkle_ids.MakeWriterWithRepool(env, nil)
}

func (env env) MakeWriteDigester() interfaces.WriteBlobIdGetter {
	return merkle_ids.MakeWriter(env, nil)
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
	digest, err := Env.MakeDigestFromHash(hash)
	errors.PanicIfError(err)

	return digest
}
