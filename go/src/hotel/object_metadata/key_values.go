package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type keyValues struct {
	SelfWithoutTai markl.Id // TODO move to a separate key-value store
}

func (index *Index) GetSelfWithoutTai() interfaces.MarklId {
	return &index.SelfWithoutTai
}

func (index *Index) GetSelfWithoutTaiMutable() interfaces.MutableMarklId {
	return &index.SelfWithoutTai
}
