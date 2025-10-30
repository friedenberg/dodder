package queries

import "code.linenisgreat.com/dodder/go/src/echo/ids"

type pinnedObjectId struct {
	ids.Sigil
	ObjectId
}
