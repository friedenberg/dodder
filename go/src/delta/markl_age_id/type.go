package markl_age_id

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
)

type tipe struct{}

var _ interfaces.MarklType = tipe{}

func (tipe tipe) GetMarklTypeId() string {
	return markl.TypeIdAgeX25519Sec
}
