package pool

import "code.linenisgreat.com/dodder/go/src/_/interfaces"

// type typeRewritten[FROM TO, TO interface{}] struct {
// 	inner interfaces.PoolValue[FROM]
// }

// func MakeTypeRewrittenPool[TO any, FROM interface{ TO }](p interfaces.PoolValue[FROM]) typeRewritten[FROM, TO] {
// 	return
// }

type Bespoke[T any] struct {
	FuncGet func() T
	FuncPut func(T)
}

func (ip Bespoke[T]) Get() T {
	return ip.FuncGet()
}

func (pool Bespoke[SWIMMER]) GetWithRepool() (SWIMMER, interfaces.FuncRepool) {
	element := pool.Get()

	return element, func() {
		pool.Put(element)
	}
}

func (ip Bespoke[T]) Put(i T) (err error) {
	ip.FuncPut(i)
	return err
}
