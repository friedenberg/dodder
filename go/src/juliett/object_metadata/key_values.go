package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type keyValues struct {
	SelfWithoutTai markl.Id // TODO move to a separate key-value store
}

func (metadata *metadata) GetSelfWithoutTai() interfaces.MarklId {
	return &metadata.Index.SelfWithoutTai
}

func (metadata *metadata) GetSelfWithoutTaiMutable() interfaces.MutableMarklId {
	return &metadata.Index.SelfWithoutTai
}
