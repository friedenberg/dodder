package store_config

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
)

type mutable_config_blob struct {
	repo_configs.TypedBlob
}

func (store *store) loadMutableConfigBlob(
	mutableConfigType ids.Type,
	blobSha interfaces.Sha,
) (err error) {
	var readCloser io.ReadCloser

	if readCloser, err = store.envRepo.GetDefaultBlobStore().BlobReader(
		blobSha,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, readCloser)

	typedBlob := repo_configs.TypedBlob{
		Type: mutableConfigType,
	}

	if _, err = repo_configs.Coder.DecodeFrom(
		&typedBlob,
		readCloser,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	store.config.configMutableBlob = typedBlob.Blob

	return
}
