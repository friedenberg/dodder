package markl

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
)

var idPool interfaces.Pool[Id, *Id] = pool.MakeWithResetable[Id]()

func GetId() interfaces.MutableMarklId {
	return idPool.Get()
}

func PutId(id interfaces.MarklId) {
	switch id := id.(type) {
	case Id:
		idPool.Put(&id)

	case *Id:
		idPool.Put(id)

	default:
		panic(errors.Errorf("unsupported id type: %T", id))
	}
}
