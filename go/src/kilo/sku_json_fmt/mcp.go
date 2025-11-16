package sku_json_fmt

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type MCP struct {
	WithDormant

	URI         string   `json:"uri,omitempty"`
	RelatedURIs []string `json:"related_uris,omitempty"`
}

func (json *MCP) FromStringAndMetadata(
	objectId string,
	metadata object_metadata.IMetadataMutable,
	blobStore interfaces.BlobStore,
) (err error) {
	if err = json.Transacted.FromStringAndMetadata(
		objectId,
		metadata,
		blobStore,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	json.URI = fmt.Sprintf("dodder:///objects/%s", objectId)

	json.RelatedURIs = make([]string, 0, 2+len(json.Tags))

	json.RelatedURIs = append(
		json.RelatedURIs,
		fmt.Sprintf("dodder:///objects/%s", metadata.GetType()),
	)

	for _, tag := range json.Tags {
		json.RelatedURIs = append(
			json.RelatedURIs,
			fmt.Sprintf("dodder:///objects/%s", tag),
		)
	}

	// json.MCPURI = fmt.Sprintf(
	// 	"dodder://objects/%s@%s:%s",
	// 	objectId,
	// 	json.RepoPubkey,
	// 	json.RepoSig,
	// )

	return err
}

func (json *MCP) FromTransacted(
	object *sku.Transacted,
	blobStore interfaces.BlobStore,
) (err error) {
	return json.FromStringAndMetadata(
		object.ObjectId.String(),
		object.GetMetadataMutable(),
		blobStore,
	)
}

func (json *MCP) ToTransacted(
	object *sku.Transacted,
	blobStore interfaces.BlobStore,
) (err error) {
	if err = json.Transacted.ToTransacted(object, blobStore); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
