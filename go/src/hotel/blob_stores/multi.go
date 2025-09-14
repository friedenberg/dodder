package blob_stores

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

type Multi struct {
	ctx         interfaces.ActiveContext
	childStores []BlobStoreInitialized
}

var _ interfaces.BlobAccess = Multi{}

func (parentStore Multi) HasBlob(id interfaces.MarklId) bool {
	for _, childStore := range parentStore.childStores {
		if childStore.HasBlob(id) {
			return true
		}
	}

	return false
}

func (parentStore Multi) MakeBlobReader(
	id interfaces.MarklId,
) (interfaces.BlobReader, error) {
	for _, childStore := range parentStore.childStores {
		if childStore.HasBlob(id) {
			return childStore.MakeBlobReader(id)
		}
	}

	return nil, env_dir.ErrBlobMissing{
		BlobId: markl.Clone(id),
	}
}

func (parentStore Multi) MakeBlobWriter(
	marklHashType interfaces.FormatHash,
) (interfaces.BlobWriter, error) {
	writers := make([]io.Writer, len(parentStore.childStores))

	multiWriter := multiStoreBlobWriter{
		blobWriters: make(
			[]interfaces.BlobWriter,
			len(parentStore.childStores),
		),
	}

	for i, childStore := range parentStore.childStores {
		var err error

		if multiWriter.blobWriters[i], err = childStore.MakeBlobWriter(
			marklHashType,
		); err != nil {
			err = errors.Wrap(err)
			return nil, err
		}

		writers[i] = multiWriter.blobWriters[i]
	}

	multiWriter.Writer = io.MultiWriter(writers...)

	return multiWriter, nil
}

type multiStoreBlobWriter struct {
	io.Writer
	blobWriters []interfaces.BlobWriter
}

var _ interfaces.BlobWriter = multiStoreBlobWriter{}

func (parentWriter multiStoreBlobWriter) ReadFrom(
	reader io.Reader,
) (n int64, err error) {
	if n, err = io.Copy(parentWriter, reader); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (parentWriter multiStoreBlobWriter) Close() error {
	for _, childWriter := range parentWriter.blobWriters {
		if err := childWriter.Close(); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return nil
}

func (parentWriter multiStoreBlobWriter) GetMarklId() (first interfaces.MarklId) {
	for _, childWriter := range parentWriter.blobWriters {
		next := childWriter.GetMarklId()

		if first == nil {
			first = next
		} else if err := markl.AssertEqual(first, next); err != nil {
			panic(err)
		}
	}

	return
}
