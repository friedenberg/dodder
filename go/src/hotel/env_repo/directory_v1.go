package env_repo

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

type directoryV1 struct {
	storeVersion interfaces.StoreVersion
	xdg.XDG
}

func (c *directoryV1) init(
	storeVersion interfaces.StoreVersion,
	xdg xdg.XDG,
) (err error) {
	c.storeVersion = storeVersion
	c.XDG = xdg
	return
}

func (c directoryV1) GetDirectoryPaths() interfaces.DirectoryPaths {
	return c
}

func (c directoryV1) FileCacheDormant() string {
	return c.DirDodder("dormant")
}

func (c directoryV1) FileTags() string {
	return c.DirDodder("tags")
}

func (c directoryV1) FileLock() string {
	return filepath.Join(c.State, "lock")
}

func (c directoryV1) FileConfigPermanent() string {
	return c.DirDodder("config-permanent")
}

func (c directoryV1) FileConfigMutable() string {
	return c.DirDodder("config-mutable")
}

func (s directoryV1) Dir(p ...string) string {
	return filepath.Join(stringSliceJoin(s.Data, p)...)
}

func (s directoryV1) DirDodder(p ...string) string {
	return s.Dir(p...)
}

func (s directoryV1) DirCache(p ...string) string {
	return s.DirDodder(append([]string{"cache"}, p...)...)
}

func (s directoryV1) DirCacheRepo(p ...string) string {
	// TODO switch to XDG cache
	// return filepath.Join(stringSliceJoin(s.Cache, "repo", p...)...)
	return s.DirDodder(append([]string{"cache", "repo"}, p...)...)
}

func (s directoryV1) DirBlobStores(p ...string) string {
	return s.DirDodder(append([]string{"objects"}, p...)...)
}

func (s directoryV1) DirBlobStoreConfigs(p ...string) string {
	return s.DirBlobStores(p...)
}

func (s directoryV1) DirLostAndFound() string {
	return s.DirDodder("lost_and_found")
}

func (s directoryV1) DirCacheObjects() string {
	return s.DirCache("objects")
}

func (s directoryV1) DirCacheObjectPointers() string {
	return s.DirCache("object_pointers")
}

func (s directoryV1) DirCacheInventoryListLog() string {
	return s.DirCache("inventory_list_logs")
}

func (s directoryV1) DirObjectId() string {
	return s.DirDodder("object_ids")
}

func (s directoryV1) FileCacheObjectId() string {
	return s.DirCache("object_id")
}

func (s directoryV1) FileInventoryListLog() string {
	return s.DirBlobStores("inventory_lists_log")
}

func (s directoryV1) DirFirstBlobStoreInventoryLists() string {
	return s.DirBlobStores("inventory_lists")
}

func (s directoryV1) DirFirstBlobStoreBlobs() string {
	return s.DirBlobStores("blobs")
}
