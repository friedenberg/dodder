package queries

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
)

type HoistedId = interfaces.ObjectId

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
