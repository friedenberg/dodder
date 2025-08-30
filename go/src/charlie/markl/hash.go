package markl

import (
	"hash"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type Hash struct {
	hash     hash.Hash
	hashType *HashType
	written  int64
}

var _ interfaces.Hash = &Hash{}

func (hash *Hash) Write(bites []byte) (written int, err error) {
	written, err = hash.hash.Write(bites)
	hash.written += int64(written)
	return written, err
}

func (hash *Hash) Sum(bites []byte) []byte {
	return hash.hash.Sum(bites)
}

func (hash *Hash) Reset() {
	hash.written = 0
	hash.hash.Reset()
}

func (hash *Hash) Size() int {
	return hash.hash.Size()
}

func (hash *Hash) BlockSize() int {
	return hash.hash.BlockSize()
}

func (hash *Hash) GetMarklType() interfaces.MarklType {
	return hash.hashType
}

func (hash *Hash) GetMarklId() (interfaces.MutableMarklId, interfaces.FuncRepool) {
	id := idPool.Get()
	id.tipe = hash.GetMarklType()
	id.allocDataIfNecessary(hash.Size())

	if hash.written > 0 {
		// TODO verify this works as expected
		id.data = hash.hash.Sum(id.data)
	}

	return id, func() {
		idPool.Put(id)
	}
}

func (hash *Hash) GetBlobIdForReader(
	reader io.Reader,
) (interfaces.MarklId, interfaces.FuncRepool) {
	id := idPool.Get()
	id.tipe = hash.GetMarklType()
	id.allocDataAndSetToCapIfNecessary(hash.Size())

	if _, err := io.ReadFull(reader, id.data); err != nil && err != io.EOF {
		panic(errors.Wrap(err))
	}

	return id, func() {
		idPool.Put(id)
	}
}

func (hash *Hash) GetBlobIdForReaderAt(
	reader io.ReaderAt,
	off int64,
) (interfaces.MarklId, interfaces.FuncRepool) {
	id := idPool.Get()
	id.tipe = hash.GetMarklType()
	id.allocDataAndSetToCapIfNecessary(hash.Size())

	if _, err := reader.ReadAt(id.data, off); err != nil {
		panic(errors.Wrap(err))
	}

	return id, func() {
		idPool.Put(id)
	}
}
