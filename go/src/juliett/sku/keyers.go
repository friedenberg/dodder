package sku

import (
	"encoding/gob"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

var (
	transactedKeyerObjectId   ObjectIdKeyer[*Transacted]
	externalLikeKeyerObjectId = GetExternalLikeKeyer[ExternalLike]()
	CheckedOutKeyerObjectId   = GetExternalLikeKeyer[*CheckedOut]()
)

func init() {
	gob.Register(transactedKeyerObjectId)
}

func GetExternalLikeKeyer[
	T interface {
		ExternalObjectIdGetter
		ids.ObjectIdGetter
		ExternalLikeGetter
	},
]() interfaces.StringKeyer[T] {
	return interfaces.CompoundKeyer[T]{
		ObjectIdKeyer[T]{},
		ExternalObjectIdKeyer[T]{},
		DescriptionKeyer[T]{},
	}
}

type ObjectIdKeyer[T ids.ObjectIdGetter] struct{}

func (sk ObjectIdKeyer[T]) GetKey(e T) (key string) {
	if e.GetObjectId().IsEmpty() {
		return
	}

	key = e.GetObjectId().String()

	return
}

type ExternalObjectIdKeyer[T ExternalObjectIdGetter] struct{}

func (ExternalObjectIdKeyer[T]) GetKey(e T) (key string) {
	if e.GetExternalObjectId().IsEmpty() {
		return
	}

	key = e.GetExternalObjectId().String()

	return
}

type DescriptionKeyer[T ExternalLikeGetter] struct{}

func (DescriptionKeyer[T]) GetKey(el T) (key string) {
	key = el.GetSkuExternal().Metadata.Description.String()
	return
}
