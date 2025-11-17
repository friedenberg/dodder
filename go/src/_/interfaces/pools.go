package interfaces

// an object borrowed from a pool that knows how to return itself
// type Borrowed interface {
// 	Return()
// }

type FuncRepool func()

type PoolablePtr[SWIMMER any] interface {
	Ptr[SWIMMER]
}

type PoolValue[SWIMMER any] interface {
	Get() SWIMMER
	GetWithRepool() (SWIMMER, FuncRepool)
	Put(i SWIMMER) (err error)
}

type Pool[SWIMMER any, SWIMMER_PTR Ptr[SWIMMER]] interface {
	PoolValue[SWIMMER_PTR]
}

// TODO remove below in favor of panicking

type PoolWithErrors[SWIMMER any] interface {
	Get() (SWIMMER, error)
	Put(i SWIMMER) (err error)
}

type PoolWithErrorsPtr[SWIMMER any, SWIMMER_PTR Ptr[SWIMMER]] interface {
	PoolWithErrors[SWIMMER_PTR]
}
