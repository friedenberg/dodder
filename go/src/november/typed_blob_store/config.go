package typed_blob_store

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/toml"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/repo_configs"
	"code.linenisgreat.com/dodder/go/src/kilo/env_repo"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
	"code.linenisgreat.com/dodder/go/src/lima/blob_library"
)

type Config struct {
	toml_v0 interfaces.TypedStore[repo_configs.V0, *repo_configs.V0]
	toml_v1 interfaces.TypedStore[repo_configs.V1, *repo_configs.V1]
}

func MakeConfigStore(
	envRepo env_repo.Env,
) Config {
	return Config{
		toml_v0: blob_library.MakeBlobStore(
			envRepo,
			blob_library.MakeBlobFormat(
				toml.MakeTomlDecoderIgnoreTomlErrors[repo_configs.V0](
					envRepo.GetDefaultBlobStore(),
				),
				toml.TomlBlobEncoder[repo_configs.V0, *repo_configs.V0]{},
				envRepo.GetDefaultBlobStore(),
			),
			func(a *repo_configs.V0) {
				a.Reset()
			},
		),
		toml_v1: blob_library.MakeBlobStore(
			envRepo,
			blob_library.MakeBlobFormat(
				toml.MakeTomlDecoderIgnoreTomlErrors[repo_configs.V1](
					envRepo.GetDefaultBlobStore(),
				),
				toml.TomlBlobEncoder[repo_configs.V1, *repo_configs.V1]{},
				envRepo.GetDefaultBlobStore(),
			),
			func(a *repo_configs.V1) {
				a.Reset()
			},
		),
	}
}

func (a Config) ParseTypedBlob(
	tipe ids.IType,
	blobId interfaces.MarklId,
) (common repo_configs.ConfigOverlay, repool interfaces.FuncRepool, n int64, err error) {
	switch tipe.String() {
	case "", ids.TypeTomlConfigV0:
		store := a.toml_v0
		var blob *repo_configs.V0

		if blob, repool, err = store.GetBlob(blobId); err != nil {
			err = errors.Wrap(err)
			return common, repool, n, err
		}

		common = blob

	case ids.TypeTomlConfigV1:
		store := a.toml_v1
		var blob *repo_configs.V1

		if blob, repool, err = store.GetBlob(blobId); err != nil {
			err = errors.Wrap(err)
			return common, repool, n, err
		}

		common = blob
	}

	return common, repool, n, err
}

func (a Config) FormatTypedBlob(
	objectGetter sku.TransactedGetter,
	writer io.Writer,
) (n int64, err error) {
	object := objectGetter.GetSku()

	tipe := object.GetType()
	blobSha := object.GetBlobDigest()

	var store interfaces.SavedBlobFormatter
	switch tipe.String() {
	case "", ids.TypeTomlConfigV0:
		store = a.toml_v0

	case ids.TypeTomlConfigV1:
		store = a.toml_v1
	}

	if n, err = store.FormatSavedBlob(writer, blobSha); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}
