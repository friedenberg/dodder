package env_repo

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

// TODO examine the directories and use XDG more appropriately for them
type directoryV2 struct {
	storeVersoin interfaces.StoreVersion
	xdg          xdg.XDG
}

func (c *directoryV2) init(
	sv interfaces.StoreVersion,
	xdg xdg.XDG,
) (err error) {
	c.storeVersoin = sv
	c.xdg = xdg
	return
}

func (c directoryV2) GetDirectoryPaths() interfaces.DirectoryPaths {
	return c
}

func (c directoryV2) FileCacheDormant() string {
	return c.DirDodder("dormant")
}

func (c directoryV2) FileTags() string {
	return c.DirDodder("tags")
}

func (c directoryV2) FileLock() string {
	return filepath.Join(c.xdg.State, "lock")
}

func (c directoryV2) FileConfigPermanent() string {
	return c.DirDodder("config-permanent")
}

func (c directoryV2) FileConfigMutable() string {
	return c.DirDodder("config-mutable")
}

func (s directoryV2) Dir(p ...string) string {
	return filepath.Join(stringSliceJoin(s.xdg.Data, p)...)
}

func (s directoryV2) DirDodder(p ...string) string {
	return s.Dir(p...)
}

func (s directoryV2) DirCache(p ...string) string {
	return s.DirDodder(append([]string{"cache"}, p...)...)
}

func (s directoryV2) DirCacheRepo(p ...string) string {
	// TODO switch to XDG cache
	// return filepath.Join(stringSliceJoin(s.Cache, "repo", p...)...)
	return s.DirDodder(append([]string{"cache", "repo"}, p...)...)
}

func (s directoryV2) DirObjects(p ...string) string {
	return s.DirDodder(append([]string{"objects"}, p...)...)
}

func (s directoryV2) DirLostAndFound() string {
	return s.DirDodder("lost_and_found")
}

func (s directoryV2) DirCacheObjects() string {
	return s.DirCache("objects")
}

func (s directoryV2) DirCacheObjectPointers() string {
	return s.DirCache("object_pointers")
}

func (s directoryV2) DirCacheInventoryListLog() string {
	return s.DirCache("inventory_list_logs")
}

func (s directoryV2) DirObjectId() string {
	return s.DirDodder("object_ids")
}

func (s directoryV2) FileCacheObjectId() string {
	return s.DirCache("object_id")
}

func (s directoryV2) FileInventoryListLog() string {
	return s.DirObjects("inventory_lists_log")
}

func (s directoryV2) DirInventoryLists() string {
	return s.DirObjects("inventory_lists")
}

// TODO switch to named blob stores
func (s directoryV2) DirBlobs() string {
	return s.DirObjects("blobs")
}
