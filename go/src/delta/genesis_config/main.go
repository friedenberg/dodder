package genesis_config

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_config"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

const (
	InventoryListTypeV0       = "!inventory_list-v0"
	InventoryListTypeV1       = "!inventory_list-v1"
	InventoryListTypeV2       = "!inventory_list-v2"
	InventoryListTypeVCurrent = InventoryListTypeV2
)

type (
	public  struct{}
	private struct{}
)

type common interface {
	GetImmutableConfigPublic() Public
	GetStoreVersion() StoreVersion
	GetPublicKey() repo_signing.PublicKey
	GetRepoType() repo_type.Type
	GetRepoId() ids.RepoId
	GetInventoryListTypeString() string

	// TODO extricate
	GetBlobStoreConfigImmutable() interfaces.BlobStoreConfigImmutable
}

// switch public and private to be "views" on the underlying interface
type Public interface {
	config() public
	common
}

type Private interface {
	common
	config() private
	GetImmutableConfig() Private
	GetPrivateKey() repo_signing.PrivateKey
}

type PrivateMutable interface {
	Private
	// TODO separate into non-method function that uses properties
	SetFlagSet(*flag.FlagSet)
	SetRepoType(repo_type.Type)
	SetRepoId(ids.RepoId)
	repo_signing.Generator
}

func DefaultMutable() PrivateMutable {
	return DefaultMutableWithVersion(store_version.VCurrent)
}

func DefaultMutableWithVersion(storeVersion StoreVersion) PrivateMutable {
	if store_version.LessOrEqual(storeVersion, store_version.V10) || true {
		return &TomlV1Private{
			TomlV1Common: TomlV1Common{
				StoreVersion:      storeVersion,
				RepoType:          repo_type.TypeWorkingCopy,
				BlobStore:         blob_store_config.Default(),
				InventoryListType: InventoryListTypeV2,
			},
		}
	} else {
		return &TomlV2Private{
			TomlV2Common: TomlV2Common{
				StoreVersion:      storeVersion,
				RepoType:          repo_type.TypeWorkingCopy,
				InventoryListType: InventoryListTypeV2,
			},
		}
	}
}
