package sku

import (
	"code.linenisgreat.com/dodder/go/src/_/external_state"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/hotel/object_metadata"
)

var TransactedResetter transactedResetter

type transactedResetter struct{}

func (transactedResetter) Reset(dst *Transacted) {
	dst.ObjectId.Reset()
	object_metadata.Resetter.Reset(&dst.Metadata)
	dst.ExternalType.Reset()
	dst.RepoId.Reset()
	dst.State = external_state.Unknown
}

func (transactedResetter) ResetWith(dst *Transacted, src *Transacted) {
	errors.PanicIfError(dst.ObjectId.ResetWithIdLike(&src.ObjectId))
	object_metadata.Resetter.ResetWith(&dst.Metadata, &src.Metadata)
	dst.ExternalType = src.ExternalType
	dst.RepoId = src.RepoId
	dst.State = src.State
	dst.ExternalObjectId.ResetWith(&src.ExternalObjectId)
}

func (transactedResetter) ResetWithExceptFields(dst *Transacted, src *Transacted) {
	errors.PanicIfError(dst.ObjectId.ResetWithIdLike(&src.ObjectId))
	object_metadata.Resetter.ResetWithExceptFields(&dst.Metadata, &src.Metadata)
	dst.ExternalType = src.ExternalType
	dst.RepoId = src.RepoId
	dst.State = src.State
	dst.ExternalObjectId.ResetWith(&src.ExternalObjectId)
}

var Resetter resetter

type resetter struct{}

func (resetter) Reset(sl TransactedGetter) {
	TransactedResetter.Reset(sl.GetSku())
}

func (resetter) ResetWith(dst, src TransactedGetter) {
	TransactedResetter.ResetWith(dst.GetSku(), src.GetSku())
}

var CheckedOutResetter checkedOutResetter

type checkedOutResetter struct{}

func (checkedOutResetter) Reset(dst *CheckedOut) {
	TransactedResetter.Reset(dst.GetSku())
	TransactedResetter.Reset(dst.GetSkuExternal().GetSku())
	dst.SetState(checked_out_state.Unknown)
}

func (checkedOutResetter) ResetWith(dst *CheckedOut, src *CheckedOut) {
	TransactedResetter.ResetWith(dst.GetSku(), src.GetSku())
	TransactedResetter.ResetWith(dst.GetSkuExternal().GetSku(), src.GetSkuExternal().GetSku())
	dst.SetState(src.state)
}
