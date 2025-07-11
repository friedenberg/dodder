package typed_blob_store

import (
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type ApproximatedType struct {
	HasValue bool
	IsActual bool
	Type     *sku.Transacted
}

func (a ApproximatedType) ActualOrNil() (actual *sku.Transacted) {
	if a.HasValue && a.IsActual {
		actual = a.Type
	}

	return
}

func (a ApproximatedType) ApproximatedOrActual() *sku.Transacted {
	if !a.HasValue {
		return nil
	}

	return a.Type
}
