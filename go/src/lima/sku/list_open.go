package sku

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/foxtrot/descriptions"
)

// TODO add lock
type OpenList struct {
	description descriptions.Description

	format ListCoder
	mover  interfaces.BlobWriter
	count  int

	funcPreWrite func(*Transacted) error
}

func MakeOpenList(
	format ListCoder,
	mover interfaces.BlobWriter,
	funcPreWrite interfaces.FuncIter[*Transacted],
) *OpenList {
	return &OpenList{
		format:       format,
		mover:        mover,
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

func (list *OpenList) Add(object *Transacted) (n int64, err error) {
	if list.funcPreWrite != nil {
		if err = list.funcPreWrite(object); err != nil {
			err = errors.Wrap(err)
			return n, err
		}
	}

	if n, err = list.writeObject(object); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return
}

// TODO swap this to not overwrite, as when importing from remotes, we want to
// keep their signatures
func (list *OpenList) writeObject(
	object *Transacted,
) (n int64, err error) {
	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(list.mover)
	defer repoolBufferedWriter()

	if n, err = list.format.EncodeTo(
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
	if err = list.mover.Close(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return
}

func (list *OpenList) GetMarklId() interfaces.MarklId {
	return list.mover.GetMarklId()
}
