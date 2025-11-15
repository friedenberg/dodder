package markl

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var idPool interfaces.Pool[Id, *Id] = pool.MakeWithResetable[Id]()

func PutBlobId(digest interfaces.MarklId) {
	switch id := digest.(type) {
	case Id:
		idPool.Put(&id)

	case *Id:
		idPool.Put(id)

	default:
		panic(errors.Errorf("unsupported id type: %T", digest))
	}
}
