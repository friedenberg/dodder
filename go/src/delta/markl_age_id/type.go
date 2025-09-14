package markl_age_id

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
)

type tipe struct{}

var _ interfaces.MarklFormat = tipe{}

func (tipe tipe) GetMarklFormatId() string {
	return markl.TypeIdAgeX25519Sec
}
