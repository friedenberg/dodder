package sku

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/delta/ohio"
	"code.linenisgreat.com/dodder/go/src/foxtrot/descriptions"
)

type ListCoder = interfaces.CoderBufferedReadWriter[*Transacted]

// TODO add lock
// TODO add iterate method
type OpenList struct {
	description descriptions.Description

	coder      ListCoder
	blobWriter interfaces.BlobWriter
	cursor     ohio.Cursor
	count      int

	funcPreWrite func(*Transacted) error
}

func MakeOpenList(
	coder ListCoder,
	blobWriter interfaces.BlobWriter,
	funcPreWrite interfaces.FuncIter[*Transacted],
) *OpenList {
	return &OpenList{
		coder:        coder,
		blobWriter:   blobWriter,
		funcPreWrite: funcPreWrite,
	}
}

func (list *OpenList) GetDescription() descriptions.Description {
	return list.description
}

func (list *OpenList) GetDescriptionMutable() *descriptions.Description {
	return &list.description
}

func (list *OpenList) Len() int {
	return list.count
}

func (list *OpenList) Add(object *Transacted) (err error) {
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

	return err
}

func (list *OpenList) writeObject(
	object *Transacted,
) (n int64, err error) {
	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(list.blobWriter)
	defer repoolBufferedWriter()

	if n, err = list.coder.EncodeTo(
		object,
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	if err = bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	list.count += 1

	return n, err
}

func (list *OpenList) Close() (err error) {
	if err = list.blobWriter.Close(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	list.cursor.Reset()

	return err
}

func (list *OpenList) GetMarklId() interfaces.MarklId {
	return list.blobWriter.GetMarklId()
}
