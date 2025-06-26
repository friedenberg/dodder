package pool

import "code.linenisgreat.com/dodder/go/src/alfa/interfaces"

type fakePool[T any, TPtr interfaces.Ptr[T]] struct{}

func MakeFakePool[T any, TPtr interfaces.Ptr[T]]() *fakePool[T, TPtr] {
	return &fakePool[T, TPtr]{}
}

func (ip fakePool[T, TPtr]) Get() TPtr {
	var t T
	return &t
}

func (ip fakePool[T, TPtr]) Put(i TPtr) (err error) {
	return
}

func (ip fakePool[T, TPtr]) PutMany(is ...TPtr) (err error) {
	return
}
