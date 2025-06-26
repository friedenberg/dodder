package sku_fmt

import (
	"bytes"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
)

type formatterTypFormatterUTIGroups struct {
	sku.OneReader
	typed_blob_store.Type
}

func MakeFormatterTypFormatterUTIGroups(
	sr sku.OneReader,
	typeBlobStore typed_blob_store.Type,
) *formatterTypFormatterUTIGroups {
	return &formatterTypFormatterUTIGroups{
		OneReader: sr,
		Type:      typeBlobStore,
	}
}

func (e formatterTypFormatterUTIGroups) Format(
	w io.Writer,
	z *sku.Transacted,
) (n int64, err error) {
	var skuTyp *sku.Transacted

	if skuTyp, err = e.ReadTransactedFromObjectId(z.Metadata.GetType()); err != nil {
		err = errors.Wrap(err)
		return
	}

	var blob type_blobs.Blob

	if blob, _, err = e.ParseTypedBlob(
		skuTyp.GetType(),
		skuTyp.GetBlobSha(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer e.PutTypedBlob(skuTyp.GetType(), blob)

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

		if n, err = io.Copy(w, sb); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
