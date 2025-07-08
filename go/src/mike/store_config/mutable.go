package store_config

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/repo_config_blobs"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
)

type mutable_config_blob struct {
	typedConfigBlobStore typed_blob_store.Config
	repo_config_blobs.Blob
}

func (k *mutable_config_blob) loadMutableConfigBlob(
	mutableConfigType ids.Type,
	blobSha interfaces.Sha,
) (err error) {
	if k.Blob, _, err = k.typedConfigBlobStore.ParseTypedBlob(
		mutableConfigType,
		blobSha,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
