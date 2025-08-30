package markl

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var idPool interfaces.Pool[Id, *Id] = pool.MakeWithResetable[Id]()

func PutBlobId(digest interfaces.BlobId) {
	switch id := digest.(type) {
	case Id:
		idPool.Put(&id)

	case *Id:
		idPool.Put(id)

	default:
		tipe := digest.GetType()

		if env, ok := envs[tipe]; ok {
			env.PutBlobId(digest)
		} else {
			panic(errors.Errorf("no env registered for digest type: %s", tipe))
		}
	}
}
