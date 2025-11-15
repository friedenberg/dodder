package ids

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
)

type ProbeId struct {
	Key string
	Id  interfaces.MarklId
}

type ProbeIdWithObjectId struct {
	ProbeId
	ObjectId *ObjectId
}
