package genesis_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
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
		GetPublicKey() markl.PublicKey
		GetRepoType() repo_type.Type
		GetRepoId() ids.RepoId
		GetInventoryListTypeString() string
		GetObjectSigTypeString() string
		GetBlobDigestTypeString() string
	}

	ConfigPublic interface {
		Config
		GetGenesisConfig() ConfigPublic
	}

	ConfigPrivate interface {
		Config
		GetGenesisConfigPublic() ConfigPublic
		GetGenesisConfig() ConfigPrivate
		GetPrivateKey() markl.PrivateKey
	}

	ConfigPrivateMutable interface {
		ConfigPrivate

		SetInventoryListTypeString(string)
		SetObjectSigTypeString(string)
		SetBlobDigestTypeString(string)
		// TODO separate into non-method function that uses properties
		flags.CommandComponentWriter
		SetRepoType(repo_type.Type)
		SetRepoId(ids.RepoId)
		markl.Generator
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
					ObjectSigType:     markl.HRPObjectSigV1,
					BlobDigestType:    markl.HRPObjectBlobDigestSha256V0,
				},
			},
		}
	}
}
