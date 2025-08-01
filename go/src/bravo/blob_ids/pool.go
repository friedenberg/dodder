package blob_ids

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func GetBlobId(tipe string) (interfaces.BlobId, func()) {
	if env, ok := envs[tipe]; ok {
		digest := env.GetBlobId()
		return digest, func() { env.PutBlobId(digest) }
	} else {
		panic(errors.Errorf("no env registered for digest type: %s", tipe))
	}
}

func PutBlobId(digest interfaces.BlobId) {
	tipe := digest.GetType()

	if env, ok := envs[tipe]; ok {
		env.PutBlobId(digest)
	} else {
		panic(errors.Errorf("no env registered for digest type: %s", tipe))
	}
}
