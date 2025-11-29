package sku_fmt

import (
	"bytes"
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/hotel/object_metadata"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
	"code.linenisgreat.com/dodder/go/src/mike/type_blobs"
)

type TypeBlobStore interface {
	ParseTypedBlob(
		tipe interfaces.ObjectId,
		blobSha interfaces.MarklId,
	) (common type_blobs.Blob, repool interfaces.FuncRepool, n int64, err error)
}

type FuncReadTypeObject func(object_metadata.TypeLock) (*sku.Transacted, error)

type formatterTypFormatterUTIGroups struct {
	funcReadTypeObject FuncReadTypeObject
	store              TypeBlobStore
}

func MakeFormatterTypeFormatterUTIGroups(
	typeReader FuncReadTypeObject,
	typeBlobStore TypeBlobStore,
) *formatterTypFormatterUTIGroups {
	return &formatterTypFormatterUTIGroups{
		funcReadTypeObject: typeReader,
		store:              typeBlobStore,
	}
}

// TODO rewrite as coder
func (format formatterTypFormatterUTIGroups) Format(
	writer io.Writer,
	object *sku.Transacted,
) (n int64, err error) {
	var typeObject *sku.Transacted

	if typeObject, err = format.funcReadTypeObject(
		object.GetTypeLock(),
	); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	var blob type_blobs.Blob
	var repool interfaces.FuncRepool

	if blob, repool, _, err = format.store.ParseTypedBlob(
		typeObject.GetType(),
		typeObject.GetBlobDigest(),
	); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	defer repool()

	for groupName, group := range blob.GetFormatterUTIGroups() {
		sb := bytes.NewBuffer(nil)

		sb.WriteString(groupName)

		for uti, formatter := range group.Map() {
			sb.WriteString(" ")
			sb.WriteString(uti)
			sb.WriteString(" ")
			sb.WriteString(formatter)
		}

		sb.WriteString("\n")

		if n, err = io.Copy(writer, sb); err != nil {
			err = errors.Wrap(err)
			return n, err
		}
	}

	return n, err
}
