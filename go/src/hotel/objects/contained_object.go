package objects

import "code.linenisgreat.com/dodder/go/src/foxtrot/markl"

type containedObject struct {
	// TODO add path information

	// required to be exported for Gob's stupid illusions
	Lock markl.Lock[SeqId, *SeqId]
}

func (tag containedObject) GetKey() SeqId {
	return tag.Lock.GetKey()
}
