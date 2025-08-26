package merkle_ids

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func PutBlobId(digest interfaces.BlobId) {
	tipe := digest.GetType()

	if env, ok := envs[tipe]; ok {
		env.PutBlobId(digest)
	} else {
		panic(errors.Errorf("no env registered for digest type: %s", tipe))
	}
}
