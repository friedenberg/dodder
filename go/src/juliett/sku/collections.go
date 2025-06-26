package sku

import (
	"encoding/gob"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
)

var (
	TransactedSetEmpty TransactedSet
	TransactedLessor   transactedLessorStable
	TransactedEqualer  transactedEqualer
)

type Collection interfaces.Collection[*Transacted]

func init() {
	TransactedSetEmpty = MakeTransactedSet()
	gob.Register(TransactedSetEmpty)
	gob.Register(MakeTransactedMutableSet())
}

type (
	TransactedSet        = interfaces.SetLike[*Transacted]
	TransactedMutableSet = interfaces.MutableSetLike[*Transacted]

	ExternalLikeSet        = interfaces.SetLike[ExternalLike]
	ExternalLikeMutableSet = interfaces.MutableSetLike[ExternalLike]

	CheckedOutSet        = interfaces.SetLike[*CheckedOut]
	CheckedOutMutableSet = interfaces.MutableSetLike[*CheckedOut]
)

func MakeTransactedSet() TransactedSet {
	return collections_value.MakeValueSet(transactedKeyerObjectId)
}

func MakeTransactedMutableSet() TransactedMutableSet {
	return collections_value.MakeMutableValueSet(transactedKeyerObjectId)
}

func MakeExternalLikeSet() ExternalLikeSet {
	return collections_value.MakeValueSet(externalLikeKeyerObjectId)
}

func MakeExternalLikeMutableSet() ExternalLikeMutableSet {
	return collections_value.MakeMutableValueSet(externalLikeKeyerObjectId)
}

func MakeCheckedOutSet() CheckedOutSet {
	return collections_value.MakeValueSet(CheckedOutKeyerObjectId)
}

func MakeCheckedOutMutableSet() CheckedOutMutableSet {
	return collections_value.MakeMutableValueSet(CheckedOutKeyerObjectId)
}
