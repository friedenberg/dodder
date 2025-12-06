package genesis_configs

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/golf/triple_hyphen_io"
)

type (
	Config interface {
		GetStoreVersion() store_version.Version
		GetPublicKey() interfaces.MarklId
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
		GetPrivateKey() interfaces.MarklId
	}

	ConfigPrivateMutable interface {
		interfaces.CommandComponentWriter
		ConfigPrivate

		SetInventoryListTypeId(string)
		SetObjectSigMarklTypeId(string)
		SetRepoId(ids.RepoId)
		GetPrivateKeyMutable() interfaces.MarklIdMutable
	}

	TypedConfigPublic         = triple_hyphen_io.TypedBlob[ConfigPublic]
	TypedConfigPrivate        = triple_hyphen_io.TypedBlob[ConfigPrivate]
	TypedConfigPrivateMutable = triple_hyphen_io.TypedBlob[ConfigPrivateMutable]
)

func Default() *TypedConfigPrivateMutable {
	return DefaultWithVersion(
		store_version.VCurrent,
		ids.TypeInventoryListVCurrent,
	)
}

func DefaultWithVersion(
	storeVersion store_version.Version,
	inventoryListTypeString string,
) *TypedConfigPrivateMutable {
	return &TypedConfigPrivateMutable{
		Type: ids.GetOrPanic(
			ids.TypeTomlConfigImmutableV2,
		).TypeStruct,
		Blob: &TomlV2Private{
			TomlV2Common: TomlV2Common{
				StoreVersion:      storeVersion,
				InventoryListType: inventoryListTypeString,
				ObjectSigType:     markl.PurposeObjectSigV2,
			},
		},
	}
}
