package typed_blob_store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
)

type Type struct {
	toml_v0 TypedStore[type_blobs.TomlV0, *type_blobs.TomlV0]
	toml_v1 TypedStore[type_blobs.TomlV1, *type_blobs.TomlV1]
}

func MakeTypeStore(
	envRepo env_repo.Env,
) Type {
	return Type{
		toml_v0: MakeBlobStore(
			envRepo,
			MakeBlobFormat(
				MakeTomlDecoderIgnoreTomlErrors[type_blobs.TomlV0](
					envRepo.GetDefaultBlobStore(),
				),
				TomlBlobEncoder[type_blobs.TomlV0, *type_blobs.TomlV0]{},
				envRepo.GetDefaultBlobStore(),
			),
			func(a *type_blobs.TomlV0) {
				a.Reset()
			},
		),
		toml_v1: MakeBlobStore(
			envRepo,
			MakeBlobFormat(
				MakeTomlDecoderIgnoreTomlErrors[type_blobs.TomlV1](
					envRepo.GetDefaultBlobStore(),
				),
				TomlBlobEncoder[type_blobs.TomlV1, *type_blobs.TomlV1]{},
				envRepo.GetDefaultBlobStore(),
			),
			func(a *type_blobs.TomlV1) {
				a.Reset()
			},
		),
	}
}

func (store Type) GetCommonStore() interfaces.TypedBlobStore[type_blobs.Blob] {
	return store
}

// TODO return repool function
func (store Type) ParseTypedBlob(
	tipe interfaces.ObjectId,
	blobSha interfaces.BlobId,
) (common type_blobs.Blob, n int64, err error) {
	switch tipe.String() {
	default:
		err = errors.Errorf("unsupported type: %q", tipe)
		return

	case "", ids.TypeTomlTypeV0:
		store := store.toml_v0
		var blob *type_blobs.TomlV0

		if blob, err = store.GetBlob(blobSha); err != nil {
			err = errors.Wrap(err)
			return
		}

		common = blob

	case ids.TypeTomlTypeV1:
		store := store.toml_v1
		var blob *type_blobs.TomlV1

		if blob, err = store.GetBlob(blobSha); err != nil {
			err = errors.Wrap(err)
			return
		}

		common = blob
	}

	return
}

func (store Type) PutTypedBlob(
	tipe interfaces.ObjectId,
	common type_blobs.Blob,
) (err error) {
	switch tipe.String() {
	case "", ids.TypeTomlTypeV0:
		if blob, ok := common.(*type_blobs.TomlV0); !ok {
			err = errors.ErrorWithStackf("expected %T but got %T", blob, common)
			return
		} else {
			store.toml_v0.PutBlob(blob)
		}

	case ids.TypeTomlTypeV1:
		if blob, ok := common.(*type_blobs.TomlV1); !ok {
			err = errors.ErrorWithStackf("expected %T but got %T", blob, common)
			return
		} else {
			store.toml_v1.PutBlob(blob)
		}
	}

	return
}
