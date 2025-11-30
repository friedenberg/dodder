package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type keyValues struct {
	SelfWithoutTai markl.Id // TODO move to a separate key-value store
}

func (index *index) GetSelfWithoutTai() interfaces.MarklId {
	return &index.SelfWithoutTai
}

func (index *index) GetSelfWithoutTaiMutable() interfaces.MutableMarklId {
	return &index.SelfWithoutTai
}
