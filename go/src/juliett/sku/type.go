package sku

// type (
// 	SkuType           = ExternalLike
// 	SkuTypeSet        = ExternalLikeSet
// 	SkuTypeSetMutable = ExternalLikeMutableSet
// 	ObjectFactory     = objectFactoryExternalLike
// )

// var (
// 	MakeSkuType                = makeExternalLike
// 	CloneSkuType               = cloneExternalLike
// 	CloneSkuTypeFromTransacted = cloneFromTransactedExternalLike
// 	MakeSkuTypeSetMutable      = MakeExternalLikeMutableSet
// )

type (
	SkuType           = *CheckedOut
	SkuTypeSet        = CheckedOutSet
	SkuTypeSetMutable = CheckedOutMutableSet
	ObjectFactory     = objectFactoryCheckedOut
)

var (
	MakeSkuType                = makeCheckedOut
	CloneSkuType               = cloneCheckedOut
	CloneSkuTypeFromTransacted = cloneFromTransactedCheckedOut
	MakeSkuTypeSetMutable      = MakeCheckedOutMutableSet
)
