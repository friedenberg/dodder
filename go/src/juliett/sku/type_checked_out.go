package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
)

func makeCheckedOut() *CheckedOut {
	dst := GetCheckedOutPool().Get()
	return dst
}

func cloneFromTransactedCheckedOut(
	src *Transacted,
	newState checked_out_state.State,
) *CheckedOut {
	dst := GetCheckedOutPool().Get()
	TransactedResetter.ResetWith(dst.GetSku(), src)
	TransactedResetter.ResetWith(dst.GetSkuExternal(), src)
	dst.state = newState
	return dst
}

func cloneCheckedOut(co *CheckedOut) *CheckedOut {
	return co.Clone()
}

type objectFactoryCheckedOut struct {
	interfaces.PoolValue[*CheckedOut]
	interfaces.Resetter3[*CheckedOut]
}

func (factory *objectFactoryCheckedOut) SetDefaultsIfNecessary() objectFactoryCheckedOut {
	if factory.Resetter3 == nil {
		factory.Resetter3 = pool.BespokeResetter[*CheckedOut]{
			FuncReset: func(e *CheckedOut) {
				CheckedOutResetter.Reset(e)
			},
			FuncResetWith: func(dst, src *CheckedOut) {
				CheckedOutResetter.ResetWith(dst, src)
			},
		}
	}

	if factory.PoolValue == nil {
		factory.PoolValue = pool.Bespoke[*CheckedOut]{
			FuncGet: func() *CheckedOut {
				return GetCheckedOutPool().Get()
			},
			FuncPut: func(e *CheckedOut) {
				GetCheckedOutPool().Put(e)
			},
		}
	}

	return *factory
}
