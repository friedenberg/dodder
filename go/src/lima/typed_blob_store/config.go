package typed_blob_store

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type Config struct {
	toml_v0 TypedStore[repo_configs.V0, *repo_configs.V0]
	toml_v1 TypedStore[repo_configs.V1, *repo_configs.V1]
}

func MakeConfigStore(
	envRepo env_repo.Env,
) Config {
	return Config{
		toml_v0: MakeBlobStore(
			envRepo,
			MakeBlobFormat(
				MakeTomlDecoderIgnoreTomlErrors[repo_configs.V0](
					envRepo.GetDefaultBlobStore(),
				),
				TomlBlobEncoder[repo_configs.V0, *repo_configs.V0]{},
				envRepo.GetDefaultBlobStore(),
			),
			func(a *repo_configs.V0) {
				a.Reset()
			},
		),
		toml_v1: MakeBlobStore(
			envRepo,
			MakeBlobFormat(
				MakeTomlDecoderIgnoreTomlErrors[repo_configs.V1](
					envRepo.GetDefaultBlobStore(),
				),
				TomlBlobEncoder[repo_configs.V1, *repo_configs.V1]{},
				envRepo.GetDefaultBlobStore(),
			),
			func(a *repo_configs.V1) {
				a.Reset()
			},
		),
	}
}

func (a Config) ParseTypedBlob(
	tipe ids.Type,
	blobSha interfaces.BlobId,
) (common repo_configs.ConfigOverlay, n int64, err error) {
	switch tipe.String() {
	case "", ids.TypeTomlConfigV0:
		store := a.toml_v0
		var blob *repo_configs.V0

		if blob, err = store.GetBlob(blobSha); err != nil {
			err = errors.Wrap(err)
			return
		}

		common = blob

	case ids.TypeTomlConfigV1:
		store := a.toml_v1
		var blob *repo_configs.V1

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
	blobSha := sk.GetBlobDigest()

	var store interfaces.SavedBlobFormatter
	switch tipe.String() {
	case "", ids.TypeTomlConfigV0:
		store = a.toml_v0

	case ids.TypeTomlConfigV1:
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
	common repo_configs.ConfigOverlay,
) (err error) {
	switch tipe.String() {
	case "", ids.TypeTomlConfigV0:
		if blob, ok := common.(*repo_configs.V0); !ok {
			err = errors.ErrorWithStackf("expected %T but got %T", blob, common)
			return
		} else {
			a.toml_v0.PutBlob(blob)
		}

	case ids.TypeTomlConfigV1:
		if blob, ok := common.(*repo_configs.V1); !ok {
			err = errors.ErrorWithStackf("expected %T but got %T", blob, common)
			return
		} else {
			a.toml_v1.PutBlob(blob)
		}
	}

	return
}
