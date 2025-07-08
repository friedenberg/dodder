package typed_blob_store

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
	"code.linenisgreat.com/dodder/go/src/golf/repo_config_blobs"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type Config struct {
	toml_v0 TypedStore[repo_config_blobs.V0, *repo_config_blobs.V0]
	toml_v1 TypedStore[repo_config_blobs.V1, *repo_config_blobs.V1]
}

func MakeConfigStore(
	repoLayout env_repo.Env,
) Config {
	return Config{
		toml_v0: MakeBlobStore(
			repoLayout,
			MakeBlobFormat(
				MakeTomlDecoderIgnoreTomlErrors[repo_config_blobs.V0](
					repoLayout,
				),
				TomlBlobEncoder[repo_config_blobs.V0, *repo_config_blobs.V0]{},
				repoLayout,
			),
			func(a *repo_config_blobs.V0) {
				a.Reset()
			},
		),
		toml_v1: MakeBlobStore(
			repoLayout,
			MakeBlobFormat(
				MakeTomlDecoderIgnoreTomlErrors[repo_config_blobs.V1](
					repoLayout,
				),
				TomlBlobEncoder[repo_config_blobs.V1, *repo_config_blobs.V1]{},
				repoLayout,
			),
			func(a *repo_config_blobs.V1) {
				a.Reset()
			},
		),
	}
}

func (a Config) ParseTypedBlob(
	tipe ids.Type,
	blobSha interfaces.Sha,
) (common repo_config_blobs.Blob, n int64, err error) {
	switch tipe.String() {
	case "", builtin_types.ConfigTypeTomlV0:
		store := a.toml_v0
		var blob *repo_config_blobs.V0

		if blob, err = store.GetBlob(blobSha); err != nil {
			err = errors.Wrap(err)
			return
		}

		common = blob

	case builtin_types.ConfigTypeTomlV1:
		store := a.toml_v1
		var blob *repo_config_blobs.V1

		if blob, err = store.GetBlob(blobSha); err != nil {
			err = errors.Wrap(err)
			return
		}

		common = blob
	}

	return
}

func (a Config) FormatTypedBlob(
	tg sku.TransactedGetter,
	w io.Writer,
) (n int64, err error) {
	sk := tg.GetSku()

	tipe := sk.GetType()
	blobSha := sk.GetBlobSha()

	var store interfaces.SavedBlobFormatter
	switch tipe.String() {
	case "", builtin_types.ConfigTypeTomlV0:
		store = a.toml_v0

	case builtin_types.ConfigTypeTomlV1:
		store = a.toml_v1
	}

	if n, err = store.FormatSavedBlob(w, blobSha); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (a Config) PutTypedBlob(
	tipe interfaces.ObjectId,
	common repo_config_blobs.Blob,
) (err error) {
	switch tipe.String() {
	case "", builtin_types.ConfigTypeTomlV0:
		if blob, ok := common.(*repo_config_blobs.V0); !ok {
			err = errors.ErrorWithStackf("expected %T but got %T", blob, common)
			return
		} else {
			a.toml_v0.PutBlob(blob)
		}

	case builtin_types.ConfigTypeTomlV1:
		if blob, ok := common.(*repo_config_blobs.V1); !ok {
			err = errors.ErrorWithStackf("expected %T but got %T", blob, common)
			return
		} else {
			a.toml_v1.PutBlob(blob)
		}
	}

	return
}
