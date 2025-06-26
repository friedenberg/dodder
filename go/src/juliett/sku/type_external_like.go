package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

func makeExternalLike() ExternalLike {
	dst := GetTransactedPool().Get()
	return dst
}

func cloneFromTransactedExternalLike(src *Transacted) ExternalLike {
	dst := GetTransactedPool().Get()
	TransactedResetter.ResetWith(dst, src)
	return dst
}

func cloneExternalLike(el ExternalLike) ExternalLike {
	return el.GetSku().CloneTransacted()
}

type objectFactoryExternalLike struct {
	interfaces.PoolValue[ExternalLike]
	interfaces.Resetter3[ExternalLike]
}

func (of *objectFactoryExternalLike) SetDefaultsIfNecessary() objectFactoryExternalLike {
	if of.Resetter3 == nil {
		of.Resetter3 = pool.BespokeResetter[ExternalLike]{
			FuncReset: func(e ExternalLike) {
				TransactedResetter.Reset(e.GetSku())
			},
			FuncResetWith: func(dst, src ExternalLike) {
				TransactedResetter.ResetWith(dst.GetSku(), src.GetSku())
			},
		}
	}

	if of.PoolValue == nil {
		of.PoolValue = pool.Bespoke[ExternalLike]{
			FuncGet: func() ExternalLike {
				return GetTransactedPool().Get()
			},
			FuncPut: func(e ExternalLike) {
				GetTransactedPool().Put(e.(*Transacted))
			},
		}
	}

	return *of
}
