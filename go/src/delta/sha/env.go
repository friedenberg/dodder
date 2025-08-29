package sha

import (
	"crypto/sha256"
	"hash"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/pool_value"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
)

var (
	Env = env{tipe: Type}
	_   = merkle.RegisterEnv(Env)
	_   = merkle.RegisterEnv(env{merkle.HRPObjectBlobDigestSha256V1})
	_   = merkle.RegisterEnv(env{merkle.HRPObjectDigestSha256V1})

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
