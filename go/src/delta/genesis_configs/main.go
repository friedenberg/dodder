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
	ConfigPublic interface {
		GetImmutableConfigPublic() ConfigPublic
		GetStoreVersion() StoreVersion
		GetPublicKey() repo_signing.PublicKey
		GetRepoType() repo_type.Type
		GetRepoId() ids.RepoId
		GetInventoryListTypeString() string
	}

	ConfigPrivate interface {
		ConfigPublic
		GetImmutableConfig() ConfigPrivate
		GetPrivateKey() repo_signing.PrivateKey
	}

	ConfigPrivateMutable interface {
		ConfigPrivate

		// TODO separate into non-method function that uses properties
		interfaces.CommandComponent
		SetRepoType(repo_type.Type)
		SetRepoId(ids.RepoId)
		repo_signing.Generator
	}

	TypedConfigPublic         = triple_hyphen_io.TypedBlob[ConfigPublic]
	TypedConfigPrivate        = triple_hyphen_io.TypedBlob[ConfigPrivate]
	TypedConfigPrivateMutable = triple_hyphen_io.TypedBlob[ConfigPrivateMutable]
)

func Default() *TypedConfigPrivateMutable {
	return DefaultWithVersion(store_version.VCurrent)
}

func DefaultWithVersion(storeVersion StoreVersion) *TypedConfigPrivateMutable {
	if store_version.IsCurrentVersionLessOrEqualToV10() {
		return &TypedConfigPrivateMutable{
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
		return &TypedConfigPrivateMutable{
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
