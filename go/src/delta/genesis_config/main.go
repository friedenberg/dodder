package genesis_config

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_config"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

const (
	InventoryListTypeV0       = "!inventory_list-v0"
	InventoryListTypeV1       = "!inventory_list-v1"
	InventoryListTypeV2       = "!inventory_list-v2"
	InventoryListTypeVCurrent = InventoryListTypeV2
)

type LatestPrivate = TomlV1Private

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

func Default() *TomlV1Private {
	return DefaultWithVersion(store_version.VCurrent)
}

// TODO read callers of this and add blob store config there
func DefaultWithVersion(storeVersion StoreVersion) *TomlV1Private {
	return &TomlV1Private{
		TomlV1Common: TomlV1Common{
			StoreVersion: storeVersion,
			RepoType:     repo_type.TypeWorkingCopy,
			BlobStore: blob_store_config.BlobStoreTomlV1{
				CompressionType:   compression_type.CompressionTypeDefault,
				LockInternalFiles: true,
			},
			InventoryListType: InventoryListTypeV2,
		},
	}
}
