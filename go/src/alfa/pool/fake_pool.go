package pool

import "code.linenisgreat.com/dodder/go/src/_/interfaces"

type fakePool[SWIMMER any, SWIMMER_PTR interfaces.Ptr[SWIMMER]] struct{}

func MakeFakePool[
	T any,
	TPtr interfaces.Ptr[T],
]() *fakePool[T, TPtr] {
	return &fakePool[T, TPtr]{}
}

func (pool fakePool[T, TPtr]) Get() TPtr {
	var t T
	return &t
}

func (pool fakePool[T, TPtr]) Put(i TPtr) (err error) {
	return err
}

func (pool fakePool[T, TPtr]) PutMany(is ...TPtr) (err error) {
	return err
}
