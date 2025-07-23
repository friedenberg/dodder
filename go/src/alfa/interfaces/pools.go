package interfaces

// an object borrowed from a pool that knows how to return itself
type Borrowed interface {
	Return()
}

type FuncRepool func()

type Poolable[T any] any

type PoolablePtr[T any] interface {
	Ptr[T]
}

type PoolValue[T any] interface {
	Get() T
	Put(i T) (err error)
}

type Pool[T any, TPtr Ptr[T]] interface {
	PoolValue[TPtr]
	PutMany(...TPtr) error
}

type PoolWithErrors[T any] interface {
	Get() (T, error)
	Put(i T) (err error)
}

type PoolWithErrorsPtr[T any, TPtr Ptr[T]] interface {
	PoolWithErrors[TPtr]
}
