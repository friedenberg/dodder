package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/cmp"
	"code.linenisgreat.com/dodder/go/src/delta/heap"
	"code.linenisgreat.com/dodder/go/src/echo/descriptions"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type (
	Seq            = interfaces.SeqError[*Transacted]
	ListTransacted = heap.Heap[Transacted, *Transacted]

	// TODO move to inventory_list_coders
	ListCoder = interfaces.CoderBufferedReadWriter[*Transacted]

	InventoryListStore interface {
		WriteInventoryListObject(*Transacted) (err error)
		ReadLast() (max *Transacted, err error)
		AllInventoryListContents(interfaces.MarklId) Seq
		AllInventoryLists() Seq
	}

	OpenList struct {
		Tipe        ids.Type
		Mover       interfaces.BlobWriter
		Description descriptions.Description
		LastTai     ids.Tai
		Len         int
	}
)

// TODO add buffered writer
func MakeListTransacted() *ListTransacted {
	heap := heap.MakeNew(
		TransactedCompare,
		transactedResetter{},
	)

	heap.SetPool(GetTransactedPool())

	return heap
}

var ResetterList resetterList

type resetterList struct{}

func (resetterList) Reset(list *ListTransacted) {
	list.Reset()
}

func (resetterList) ResetWith(left, right *ListTransacted) {
	left.ResetWith(right)
}

func CollectList(
	seq Seq,
) (list *ListTransacted, err error) {
	list = MakeListTransacted()

	for sk, iterErr := range seq {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return list, err
		}

		list.Add(sk)
	}

	return list, err
}

type ListCheckedOut = heap.Heap[CheckedOut, *CheckedOut]

func MakeListCheckedOut() *ListCheckedOut {
	heap := heap.MakeNew(
		cmp.MakeFuncFromEqualerAndLessor3LessFirst(
			genericEqualer[*CheckedOut]{},
			genericLessorStable[*CheckedOut]{},
		),
		CheckedOutResetter,
	)

	heap.SetPool(GetCheckedOutPool())

	return heap
}

var ResetterListCheckedOut resetterListCheckedOut

type resetterListCheckedOut struct{}

func (resetterListCheckedOut) Reset(a *ListCheckedOut) {
	a.Reset()
}

func (resetterListCheckedOut) ResetWith(a, b *ListCheckedOut) {
	a.ResetWith(b)
}

func CollectListCheckedOut(
	seq interfaces.SeqError[*CheckedOut],
) (list *ListCheckedOut, err error) {
	list = MakeListCheckedOut()

	for checkedOut, iterErr := range seq {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return list, err
		}

		list.Add(checkedOut)
	}

	return list, err
}

type genericLessorTaiOnly[T ids.Clock] struct{}

func (genericLessorTaiOnly[T]) Less(a, b T) bool {
	return a.GetTai().Less(b.GetTai())
}

type clockWithObjectId interface {
	ids.Clock
	// TODO figure out common interface for this
	GetObjectId() *ids.ObjectId
}

type genericLessorStable[T clockWithObjectId] struct{}

func (genericLessorStable[T]) Less(a, b T) bool {
	if result := a.GetTai().SortCompare(b.GetTai()); !result.Equal() {
		return result.Less()
	}

	return a.GetObjectId().String() < b.GetObjectId().String()
}

type genericEqualer[T interface {
	Equals(T) bool
}] struct{}

func (genericEqualer[T]) Equals(a, b T) bool {
	return a.Equals(b)
}
