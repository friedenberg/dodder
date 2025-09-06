package genesis_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

type (
	Config interface {
		GetStoreVersion() store_version.Version
		GetPublicKey() markl.Id
		GetRepoType() repo_type.Type // TODO remove
		GetRepoId() ids.RepoId
		GetInventoryListTypeId() string
		GetObjectSigMarklTypeId() string
	}

	ConfigPublic interface {
		Config
		GetGenesisConfig() ConfigPublic
	}

	ConfigPrivate interface {
		Config
		GetGenesisConfigPublic() ConfigPublic
		GetGenesisConfig() ConfigPrivate
		GetPrivateKey() markl.Id
	}

	ConfigPrivateMutable interface {
		ConfigPrivate

		SetInventoryListTypeId(string)
		SetObjectSigMarklTypeId(string)

		// TODO separate into non-method function that uses properties
		interfaces.CommandComponentWriter
		SetRepoType(repo_type.Type) // TODO remove
		SetRepoId(ids.RepoId)
		GetPrivateKeyMutable() *markl.Id
	}

	TypedConfigPublic         = triple_hyphen_io.TypedBlob[ConfigPublic]
	TypedConfigPrivate        = triple_hyphen_io.TypedBlob[ConfigPrivate]
	TypedConfigPrivateMutable = triple_hyphen_io.TypedBlob[ConfigPrivateMutable]
)

func Default() *TypedConfigPrivateMutable {
	return DefaultWithVersion(
		store_version.VCurrent,
		ids.TypeInventoryListV2,
	)
}

func DefaultWithVersion(
	storeVersion store_version.Version,
	inventoryListTypeString string,
) *TypedConfigPrivateMutable {
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
					InventoryListType: inventoryListTypeString,
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
					InventoryListType: inventoryListTypeString,
					ObjectSigType:     markl.FormatIdObjectSigV1,
				},
			},
		}
	}
}
