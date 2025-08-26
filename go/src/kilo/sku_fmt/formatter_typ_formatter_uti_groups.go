package sku_fmt

import (
	"bytes"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type TypeBlobStore interface {
	ParseTypedBlob(
		tipe interfaces.ObjectId,
		blobSha interfaces.BlobId,
	) (common type_blobs.Blob, repool interfaces.FuncRepool, n int64, err error)
}

type formatterTypFormatterUTIGroups struct {
	sku.OneReader
	store TypeBlobStore
}

func MakeFormatterTypFormatterUTIGroups(
	oneReader sku.OneReader,
	typeBlobStore TypeBlobStore,
) *formatterTypFormatterUTIGroups {
	return &formatterTypFormatterUTIGroups{
		OneReader: oneReader,
		store:     typeBlobStore,
	}
}

// TODO rewrite as coder
func (format formatterTypFormatterUTIGroups) Format(
	writer io.Writer,
	object *sku.Transacted,
) (n int64, err error) {
	var skuTyp *sku.Transacted

	if skuTyp, err = format.ReadTransactedFromObjectId(object.Metadata.GetType()); err != nil {
		err = errors.Wrap(err)
		return
	}

	var blob type_blobs.Blob
	var repool interfaces.FuncRepool

	if blob, repool, _, err = format.store.ParseTypedBlob(
		skuTyp.GetType(),
		skuTyp.GetBlobDigest(),
	); err != nil {
		err = errors.Wrap(err)
		return
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
			return
		}
	}

	return
}
