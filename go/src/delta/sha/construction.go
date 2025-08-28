package sha

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func MustWithMerkleIdWithType(merkleId interfaces.BlobId, tipe string) *Sha {
	if digest, ok := merkleId.(*Sha); ok {
		return digest
	}

	digest := poolSha.Get()

	if !merkleId.IsNull() {
		errors.PanicIfError(
			digest.SetMerkleId(tipe, merkleId.GetBytes()),
		)
	}

	return digest
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
