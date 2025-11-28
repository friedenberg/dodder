package store_abbr

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

type indexNoAbbr[
	ID interfaces.Stringer,
	ID_PTR interfaces.SetterPtr[ID],
] struct {
	sku.IdAbbrIndexGeneric[ID, ID_PTR]
}

func (ih indexNoAbbr[ID, ID_PTR]) Abbreviate(h ID) (v string, err error) {
	v = h.String()
	return v, err
}
