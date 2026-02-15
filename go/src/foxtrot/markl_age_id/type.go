package markl_age_id

import (
	"code.linenisgreat.com/dodder/go/src/alfa/domain_interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/markl"
)

type tipe struct{}

var _ domain_interfaces.MarklFormat = tipe{}

func (tipe tipe) GetMarklFormatId() string {
	return markl.FormatIdAgeX25519Sec
}

func (tipe tipe) GetSize() int {
	panic(errors.Err501NotImplemented)
}
