package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var thePool interfaces.Pool[Metadata, *Metadata]

func init() {
	thePool = pool.MakePool[Metadata, *Metadata](
		nil,
		Resetter.Reset,
	)
}

func GetPool() interfaces.Pool[Metadata, *Metadata] {
	return thePool
}
