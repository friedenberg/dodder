package genesis_config

import (
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
	Current = TomlV1Private
	Next    = TomlV2Private
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

func Default() Current {
	return DefaultWithVersion(store_version.VCurrent)
}

// TODO read callers of this and add blob store config there
func DefaultWithVersion(storeVersion StoreVersion) Current {
	return Current{
		TomlV1Common: TomlV1Common{
			StoreVersion:      storeVersion,
			RepoType:          repo_type.TypeWorkingCopy,
			BlobStore:         blob_store_config.Default(),
			InventoryListType: InventoryListTypeV2,
		},
	}
}
