package objects

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/markl"
)

type keyValues struct {
	SelfWithoutTai markl.Id // TODO move to a separate key-value store
}

func (index *index) GetSelfWithoutTai() interfaces.MarklId {
	return &index.SelfWithoutTai
}

func (index *index) GetSelfWithoutTaiMutable() interfaces.MarklIdMutable {
	return &index.SelfWithoutTai
}
