package collections_ptr

import (
	"encoding/gob"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
)

func RegisterGobValue[T interfaces.ValueLike, TPtr interfaces.ValuePtr[T]](
	keyer interfaces.StringKeyerPtr[T, TPtr],
) {
	if keyer == nil {
		keyer = quiter.StringerKeyerPtr[T, TPtr]{}.RegisterGob()
	}

	gob.Register(keyer)

	RegisterGob[T, TPtr]()
}

func RegisterGob[T interfaces.ValueLike, TPtr interfaces.ValuePtr[T]]() {
	gob.Register(Set[T, TPtr]{})
	gob.Register(MutableSet[T, TPtr]{})
}
