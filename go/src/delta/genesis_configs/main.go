package genesis_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

type (
	Public interface {
		GetImmutableConfigPublic() Public
		GetStoreVersion() StoreVersion
		GetPublicKey() repo_signing.PublicKey
		GetRepoType() repo_type.Type
		GetRepoId() ids.RepoId
		GetInventoryListTypeString() string
	}

	Private interface {
		Public
		GetImmutableConfig() Private
		GetPrivateKey() repo_signing.PrivateKey
	}

	PrivateMutable interface {
		Private

		// TODO separate into non-method function that uses properties
		interfaces.CommandComponent
		SetRepoType(repo_type.Type)
		SetRepoId(ids.RepoId)
		repo_signing.Generator
	}

	TypedPublic         = triple_hyphen_io.TypedBlob[Public]
	TypedPrivate        = triple_hyphen_io.TypedBlob[Private]
	TypedPrivateMutable = triple_hyphen_io.TypedBlob[PrivateMutable]
)

func Default() *TypedPrivateMutable {
	return DefaultWithVersion(store_version.VCurrent)
}

func DefaultWithVersion(storeVersion StoreVersion) *TypedPrivateMutable {
	if store_version.IsCurrentVersionLessOrEqualToV10() {
		return &TypedPrivateMutable{
			Type: ids.GetOrPanic(
				ids.TypeTomlConfigImmutableV1,
			).Type,
			Blob: &TomlV1Private{
				TomlV1Common: TomlV1Common{
					StoreVersion: storeVersion,
					RepoType:     repo_type.TypeWorkingCopy,
					BlobStore: blob_store_configs.TomlV0{
						CompressionType:   compression_type.CompressionTypeDefault,
						LockInternalFiles: true,
					},
					InventoryListType: ids.TypeInventoryListV2,
				},
			},
		}
	} else {
		return &TypedPrivateMutable{
			Type: ids.GetOrPanic(
				ids.TypeTomlConfigImmutableV2,
			).Type,
			Blob: &TomlV2Private{
				TomlV2Common: TomlV2Common{
					StoreVersion:      storeVersion,
					RepoType:          repo_type.TypeWorkingCopy,
					InventoryListType: ids.TypeInventoryListV2,
				},
			},
		}
	}
}
