package genesis_configs

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

type (
	Config interface {
		GetStoreVersion() store_version.Version
		GetPublicKey() interfaces.MarklId
		GetRepoId() ids.RepoId
		GetInventoryListTypeId() string
		// TODO rename to purpose
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
		GetPrivateKey() interfaces.MarklId
	}

	ConfigPrivateMutable interface {
		ConfigPrivate

		SetInventoryListTypeId(string)
		SetObjectSigMarklTypeId(string)

		// TODO separate into non-method function that uses properties
		interfaces.CommandComponentWriter
		SetRepoId(ids.RepoId)
		GetPrivateKeyMutable() interfaces.MutableMarklId
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
	return &TypedConfigPrivateMutable{
		Type: ids.GetOrPanic(
			ids.TypeTomlConfigImmutableV2,
		).Type,
		Blob: &TomlV2Private{
			TomlV2Common: TomlV2Common{
				StoreVersion:      storeVersion,
				InventoryListType: inventoryListTypeString,
				ObjectSigType:     markl.PurposeObjectSigV1,
			},
		},
	}
}
