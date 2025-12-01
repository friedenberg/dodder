package queries

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

type HoistedId = ids.ObjectIdLike

func getStringForHoistedId(id HoistedId) string {
	switch id := id.(type) {
	case MarklId:
		return id.String()

	case ObjectId:
		return id.GetObjectId().String()

	default:
		panic(fmt.Sprintf("unsupported hoisted id: %T", id))
	}
}
