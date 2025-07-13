package genesis_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
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
	GetBlobStoreConfigImmutable() interfaces.BlobIOWrapper
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
	interfaces.CommandComponent
	SetRepoType(repo_type.Type)
	SetRepoId(ids.RepoId)
	repo_signing.Generator
}

func DefaultMutable() PrivateMutable {
	return DefaultMutableWithVersion(store_version.VCurrent)
}

func DefaultMutableWithVersion(storeVersion StoreVersion) PrivateMutable {
	if store_version.IsCurrentVersionLessOrEqualToV10() {
		return &TomlV1Private{
			TomlV1Common: TomlV1Common{
				StoreVersion: storeVersion,
				RepoType:     repo_type.TypeWorkingCopy,
				BlobStore: blob_store_configs.TomlV0{
					CompressionType:   compression_type.CompressionTypeDefault,
					LockInternalFiles: true,
				},
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
