package type_blobs

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/genres"
	"code.linenisgreat.com/dodder/go/src/charlie/toml"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/env_repo"
	"code.linenisgreat.com/dodder/go/src/kilo/blob_library"
)

type Coder struct {
	toml_v0 interfaces.TypedStore[TomlV0, *TomlV0]
	toml_v1 interfaces.TypedStore[TomlV1, *TomlV1]
}

func MakeTypeStore(
	envRepo env_repo.Env,
) Coder {
	return Coder{
		toml_v0: blob_library.MakeBlobStore(
			envRepo,
			blob_library.MakeBlobFormat(
				toml.MakeTomlDecoderIgnoreTomlErrors[TomlV0](
					envRepo.GetDefaultBlobStore(),
				),
				toml.TomlBlobEncoder[TomlV0, *TomlV0]{},
				envRepo.GetDefaultBlobStore(),
			),
			func(a *TomlV0) {
				a.Reset()
			},
		),
		toml_v1: blob_library.MakeBlobStore(
			envRepo,
			blob_library.MakeBlobFormat(
				toml.MakeTomlDecoderIgnoreTomlErrors[TomlV1](
					envRepo.GetDefaultBlobStore(),
				),
				toml.TomlBlobEncoder[TomlV1, *TomlV1]{},
				envRepo.GetDefaultBlobStore(),
			),
			func(a *TomlV1) {
				a.Reset()
			},
		),
	}
}

func (store Coder) SaveBlobText(
	tipe interfaces.ObjectId,
	blob Blob,
) (digest interfaces.MarklId, n int64, err error) {
	if err = genres.Type.AssertGenre(tipe); err != nil {
		err = errors.Wrap(err)
		return digest, n, err
	}

	switch tipe.String() {
	default:
		err = errors.Errorf("unsupported type: %q", tipe)
		return digest, n, err

	case "", ids.TypeTomlTypeV0:
		if digest, n, err = store.toml_v0.SaveBlobText(blob.(*TomlV0)); err != nil {
			err = errors.Wrap(err)
			return digest, n, err
		}

	case ids.TypeTomlTypeV1:
		if digest, n, err = store.toml_v1.SaveBlobText(blob.(*TomlV1)); err != nil {
			err = errors.Wrap(err)
			return digest, n, err
		}
	}

	return digest, n, err
}

func (store Coder) ParseTypedBlob(
	tipe interfaces.ObjectId,
	blobId interfaces.MarklId,
) (common Blob, repool interfaces.FuncRepool, n int64, err error) {
	switch tipe.String() {
	default:
		err = errors.Errorf("unsupported type: %q", tipe)
		return common, repool, n, err

	case "", ids.TypeTomlTypeV0:
		store := store.toml_v0
		var blob *TomlV0

		if blob, repool, err = store.GetBlob(blobId); err != nil {
			err = errors.Wrap(err)
			return common, repool, n, err
		}

		common = blob

	case ids.TypeTomlTypeV1:
		store := store.toml_v1
		var blob *TomlV1

		if blob, repool, err = store.GetBlob(blobId); err != nil {
			err = errors.Wrap(err)
			return common, repool, n, err
		}

		common = blob
	}

	return common, repool, n, err
}
