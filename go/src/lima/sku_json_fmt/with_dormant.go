package sku_json_fmt

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/hotel/object_metadata"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

type WithDormant struct {
	Transacted

	Dormant bool `json:"dormant"`
}

func (json *WithDormant) FromStringAndMetadata(
	objectId string,
	metadata object_metadata.IMetadataMutable,
	blobStore interfaces.BlobStore,
) (err error) {
	if err = json.Transacted.FromObjectIdStringAndMetadata(
		objectId,
		metadata,
		blobStore,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	json.Dormant = metadata.GetIndex().GetDormant().Bool()

	return err
}

func (json *WithDormant) FromTransacted(
	object *sku.Transacted,
	blobStore interfaces.BlobStore,
) (err error) {
	return json.FromStringAndMetadata(
		object.ObjectId.String(),
		object.GetMetadataMutable(),
		blobStore,
	)
}

func (json *WithDormant) ToTransacted(
	object *sku.Transacted,
	blobStore interfaces.BlobStore,
) (err error) {
	if err = json.Transacted.ToTransacted(object, blobStore); err != nil {
		err = errors.Wrap(err)
		return err
	}

	object.GetMetadataMutable().GetIndexMutable().GetDormantMutable().SetBool(json.Dormant)

	return err
}
