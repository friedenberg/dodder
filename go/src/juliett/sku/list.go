package sku

import (
	"bufio"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/heap"
	"code.linenisgreat.com/dodder/go/src/echo/descriptions"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type InventoryListStore interface {
	FormatForVersion(sv interfaces.StoreVersion) ListFormat
	WriteInventoryListObject(t *Transacted) (err error)
	ImportInventoryList(bs interfaces.BlobStore, t *Transacted) (err error)
	// WriteInventoryListStream(list *Transacted, ) (err error)
	// ReadInventoryList(ids.Tai) (*sku.Transacted, *sku.List, error)

	ReadLast() (max *Transacted, err error)

	IterInventoryList(interfaces.Digest) interfaces.SeqError[*Transacted]

	ReadAllSkus(
		f func(besty, sk *Transacted) error,
	) (err error)

	// ReadAllInventoryListsSince(
	// since ids.Tai,
	// 	f interfaces.FuncIter[*sku.Transacted],
	// ) (err error)

	IterAllInventoryLists() interfaces.SeqError[*Transacted]
	MakeImporter(ImporterOptions, StoreOptions) Importer
	ImportList(*List, Importer) error
}

// TODO refactor into being just a CoderBufferedReadWriter[*sku.Transacted]
type ListFormat interface {
	GetType() ids.Type
	GetListFormat() ListFormat
	WriteObjectToOpenList(*Transacted, *OpenList) (int64, error)
	WriteInventoryListBlob(Collection, *bufio.Writer) (int64, error)
	// TODO add context and ContinueOrPanicOnDone
	WriteInventoryListObject(*Transacted, *bufio.Writer) (int64, error)
	ReadInventoryListObject(*bufio.Reader) (int64, *Transacted, error)
	// TODO add context and ContinueOrPanicOnDone
	StreamInventoryListBlobSkus(
		rf *bufio.Reader,
		f interfaces.FuncIter[*Transacted],
	) error
}

// TODO rename to ListTransacted
type List = heap.Heap[Transacted, *Transacted]

type OpenList struct {
	Tipe        ids.Type
	Mover       interfaces.Mover
	Description descriptions.Description
	LastTai     ids.Tai
	Len         int
}

// TODO rename to MakeListTransacted
func MakeList() *List {
	h := heap.Make(
		transactedEqualer{},
		transactedLessorStable{},
		transactedResetter{},
	)

	h.SetPool(GetTransactedPool())

	return h
}

var ResetterList resetterList

type resetterList struct{}

func (resetterList) Reset(a *List) {
	a.Reset()
}

func (resetterList) ResetWith(a, b *List) {
	a.ResetWith(b)
}

func CollectList(seq interfaces.SeqError[*Transacted]) (list *List, err error) {
	list = MakeList()

	for sk, iterErr := range seq {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return
		}

		list.Add(sk)
	}

	return
}

type ListCheckedOut = heap.Heap[CheckedOut, *CheckedOut]

func MakeListCheckedOut() *ListCheckedOut {
	h := heap.Make(
		genericEqualer[*CheckedOut]{},
		genericLessorStable[*CheckedOut]{},
		CheckedOutResetter,
	)

	h.SetPool(GetCheckedOutPool())

	return h
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
			return
		}

		list.Add(checkedOut)
	}

	return
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
