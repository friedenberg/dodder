package query

import (
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type CompoundMatch struct {
	sku.Queryable
	*ObjectId
}

func (k *CompoundMatch) ContainsSku(tg sku.TransactedGetter) bool {
	if k.ObjectId.ContainsSku(tg) {
		return true
	}

	if k.Queryable.ContainsSku(tg) {
		return true
	}

	return false
}

func (k *CompoundMatch) String() string {
	return k.ObjectId.String()
}
