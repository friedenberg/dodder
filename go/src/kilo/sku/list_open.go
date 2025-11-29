package sku

import (
	"bufio"
	"fmt"
	"sync"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_map"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/heap"
	"code.linenisgreat.com/dodder/go/src/delta/ohio"
	"code.linenisgreat.com/dodder/go/src/foxtrot/descriptions"
)

type ListCoder = interfaces.CoderBufferedReadWriter[*Transacted]

// TODO add lock
// TODO add iterate method
type OpenList struct {
	lock        sync.RWMutex
	description descriptions.Description

	coder                    ListCoder
	blobWriter               interfaces.BlobWriter
	bufferedBlobWriter       *bufio.Writer
	bufferedBlobWriterRepool interfaces.FuncRepool
	cursor                   ohio.Cursor
	count                    int

	indexOrder     *heap.Heap[TransactedCursor, *TransactedCursor]
	indexObjectIds collections_map.Map[string, collections_slice.Slice[ohio.Cursor]]

	funcPreWrite func(*Transacted) error
}

func MakeOpenList(
	coder ListCoder,
	blobWriter interfaces.BlobWriter,
	funcPreWrite interfaces.FuncIter[*Transacted],
) *OpenList {
	return &OpenList{
		coder:          coder,
		blobWriter:     blobWriter,
		indexOrder:     MakeHeapTransactedCursor(),
		indexObjectIds: make(collections_map.Map[string, collections_slice.Slice[ohio.Cursor]]),
		funcPreWrite:   funcPreWrite,
	}
}

func (list *OpenList) GetDescription() descriptions.Description {
	return list.description
}

func (list *OpenList) GetDescriptionMutable() *descriptions.Description {
	return &list.description
}

func (list *OpenList) getBufferedBlobWriter() *bufio.Writer {
	if list.bufferedBlobWriter == nil {
		list.bufferedBlobWriter, list.bufferedBlobWriterRepool = pool.GetBufferedWriter(
			list.blobWriter,
		)
	}

	return list.bufferedBlobWriter
}

func (list *OpenList) Len() int {
	list.lock.RLock()
	defer list.lock.RUnlock()

	return list.count
}

func (list *OpenList) Add(object *Transacted) (err error) {
	list.lock.Lock()
	defer list.lock.Unlock()

	if list.funcPreWrite != nil {
		if err = list.funcPreWrite(object); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	list.cursor.Offset += list.cursor.ContentLength

	if list.cursor.ContentLength, err = list.writeObject(object); err != nil {
		err = errors.Wrap(err)
		return err
	}

	objectIdString := object.GetObjectId().String()

	list.indexOrder.Push(&TransactedCursor{
		tai:            object.GetTai(),
		objectIdString: objectIdString,
		cursor:         list.cursor,
	})

	{
		objects, _ := list.indexObjectIds.Get(objectIdString)
		objects.Append(list.cursor)
		list.indexObjectIds.Set(objectIdString, objects)
	}

	return err
}

func (list *OpenList) writeObject(
	object *Transacted,
) (n int64, err error) {
	if n, err = list.coder.EncodeTo(
		object,
		list.getBufferedBlobWriter(),
	); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	if err = list.getBufferedBlobWriter().Flush(); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	list.count += 1

	return n, err
}

func (list *OpenList) Close() (err error) {
	if !list.lock.TryLock() {
		err = errors.Errorf("trying to close open list while lock is acquired")
		return
	}

	defer list.lock.Unlock()

	if err = list.getBufferedBlobWriter().Flush(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = list.blobWriter.Close(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	list.cursor.Reset()
	list.indexOrder.Reset()
	list.indexObjectIds.Reset()
	list.bufferedBlobWriter = nil
	list.bufferedBlobWriterRepool()

	return err
}

func (list *OpenList) GetMarklId() interfaces.MarklId {
	if !list.lock.TryLock() {
		panic(fmt.Sprintf("trying to get markl id from open list while lock is acquired"))
	}

	defer list.lock.Unlock()

	return list.blobWriter.GetMarklId()
}
