package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var thePool interfaces.Pool[metadata, *metadata]

func init() {
	thePool = pool.Make[metadata, *metadata](
		nil,
		Resetter.Reset,
	)
}

func GetPool() interfaces.Pool[metadata, *metadata] {
	return thePool
}
