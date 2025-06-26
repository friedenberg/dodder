package interfaces

type CollectionPtr[T any, TPtr Ptr[T]] interface {
	Lenner
	IterablePtr[T, TPtr]
}

type SetPtrLike[T any, TPtr Ptr[T]] interface {
	SetLike[T]
	CollectionPtr[T, TPtr]

	GetPtr(string) (TPtr, bool)
	KeyPtr(TPtr) string

	CloneSetPtrLike() SetPtrLike[T, TPtr]
	CloneMutableSetPtrLike() MutableSetPtrLike[T, TPtr]
}

type MutableSetPtrLike[T any, TPtr Ptr[T]] interface {
	SetPtrLike[T, TPtr]
	MutableSetLike[T]
	AddPtr(TPtr) error
	DelPtr(TPtr) error
}
