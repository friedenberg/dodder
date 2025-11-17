package queries

import "code.linenisgreat.com/dodder/go/src/foxtrot/ids"

type pinnedObjectId struct {
	ids.Sigil
	ObjectId
}
