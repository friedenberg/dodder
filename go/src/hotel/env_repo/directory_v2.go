package env_repo

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

// TODO examine the directories and use XDG more appropriately for them
type directoryV2 struct {
	storeVersion interfaces.StoreVersion
	xdg          xdg.XDG
}

func (directory *directoryV2) init(
	storeVersion interfaces.StoreVersion,
	xdg xdg.XDG,
) (err error) {
	directory.storeVersion = storeVersion
	directory.xdg = xdg
	return
}

func (directory directoryV2) GetDirectoryPaths() interfaces.DirectoryPaths {
	return directory
}

func (directory directoryV2) Dir(p ...string) string {
	return filepath.Join(stringSliceJoin(directory.xdg.Data, p)...)
}

func (directory directoryV2) DirDodder(p ...string) string {
	return directory.Dir(p...)
}

func (directory directoryV2) DirBlobStores(p ...string) string {
	return directory.DirDodder(append([]string{"blob_stores"}, p...)...)
}

// TODO deprecate and remove
func (directory directoryV2) DirFirstBlobStoreInventoryLists() string {
	return directory.DirBlobStores("0/inventory_lists")
}

// TODO deprecate and remove
func (directory directoryV2) DirFirstBlobStoreBlobs() string {
	return directory.DirBlobStores("0/blobs")
}

func (directory directoryV2) FileCacheDormant() string {
	return directory.DirDodder("dormant")
}

func (directory directoryV2) FileTags() string {
	return directory.DirDodder("tags")
}

func (directory directoryV2) FileLock() string {
	return filepath.Join(directory.xdg.State, "lock")
}

func (directory directoryV2) FileConfigPermanent() string {
	return directory.DirDodder("config-permanent")
}

func (directory directoryV2) FileConfigMutable() string {
	return directory.DirDodder("config-mutable")
}

func (directory directoryV2) DirCache(p ...string) string {
	return directory.DirDodder(append([]string{"cache"}, p...)...)
}

func (directory directoryV2) DirCacheRepo(p ...string) string {
	// TODO switch to XDG cache
	// return filepath.Join(stringSliceJoin(s.Cache, "repo", p...)...)
	return directory.DirDodder(append([]string{"cache", "repo"}, p...)...)
}

func (directory directoryV2) DirLostAndFound() string {
	return directory.DirDodder("lost_and_found")
}

func (directory directoryV2) DirCacheObjects() string {
	return directory.DirCache("objects")
}

func (directory directoryV2) DirCacheObjectPointers() string {
	return directory.DirCache("object_pointers")
}

func (directory directoryV2) DirCacheInventoryListLog() string {
	return directory.DirCache("inventory_list_logs")
}

func (directory directoryV2) DirObjectId() string {
	return directory.DirDodder("object_ids")
}

func (directory directoryV2) FileCacheObjectId() string {
	return directory.DirCache("object_id")
}

func (directory directoryV2) FileInventoryListLog() string {
	return directory.DirBlobStores("inventory_lists_log")
}
