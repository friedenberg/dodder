package markl

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/domain_interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
)

var idPool interfaces.Pool[Id, *Id] = pool.MakeWithResetable[Id]()

func GetId() domain_interfaces.MarklIdMutable {
	return idPool.Get()
}

func PutId(id domain_interfaces.MarklId) {
	switch id := id.(type) {
	case Id:
		idPool.Put(&id)

	case *Id:
		idPool.Put(id)

	default:
		panic(errors.Errorf("unsupported id type: %T", id))
	}
}
