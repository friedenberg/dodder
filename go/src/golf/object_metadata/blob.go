package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/collections_slice"
)

type blob struct {
	// TODO replace with dumber field representation
	// TODO move to Cache
	Fields collections_slice.Slice[Field]
}

func (blob *blob) GetFields() interfaces.Seq[Field] {
	return blob.Fields.All()
}

func (blob *blob) GetFieldsMutable() *collections_slice.Slice[Field] {
	return &blob.Fields
}
