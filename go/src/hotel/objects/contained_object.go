package objects

import (
	"code.linenisgreat.com/dodder/go/src/alfa/cmp"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type containedObject struct {
	// TODO add path information

	// required to be exported for Gob's stupid illusions
	Lock markl.Lock[SeqId, *SeqId]
}

func (object containedObject) GetKey() SeqId {
	return object.Lock.GetKey()
}

func containedObjectCompareKey(left, right containedObject) cmp.Result {
	return SeqIdCompare(left.GetKey(), right.GetKey())
}
