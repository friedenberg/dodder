package env_repo

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

type directoryV1 struct {
	sv interfaces.StoreVersion
	xdg.XDG
}

func (c *directoryV1) init(
	sv interfaces.StoreVersion,
	xdg xdg.XDG,
) (err error) {
	c.sv = sv
	c.XDG = xdg
	return
}

func (c directoryV1) GetDirectoryPaths() interfaces.DirectoryPaths {
	return c
}

func (c directoryV1) FileCacheDormant() string {
	return c.DirZit("dormant")
}

func (c directoryV1) FileTags() string {
	return c.DirZit("tags")
}

func (c directoryV1) FileLock() string {
	return filepath.Join(c.State, "lock")
}

func (c directoryV1) FileConfigPermanent() string {
	return c.DirZit("config-permanent")
}

func (c directoryV1) FileConfigMutable() string {
	return c.DirZit("config-mutable")
}

func (s directoryV1) Dir(p ...string) string {
	return filepath.Join(stringSliceJoin(s.Data, p)...)
}

func (s directoryV1) DirZit(p ...string) string {
	return s.Dir(p...)
}

func (s directoryV1) DirCache(p ...string) string {
	return s.DirZit(append([]string{"cache"}, p...)...)
}

func (s directoryV1) DirCacheRepo(p ...string) string {
	// TODO switch to XDG cache
	// return filepath.Join(stringSliceJoin(s.Cache, "repo", p...)...)
	return s.DirZit(append([]string{"cache", "repo"}, p...)...)
}

func (s directoryV1) DirObjects(p ...string) string {
	return s.DirZit(append([]string{"objects"}, p...)...)
}

func (s directoryV1) DirLostAndFound() string {
	return s.DirZit("lost_and_found")
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
	return s.DirZit("object_ids")
}

func (s directoryV1) FileCacheObjectId() string {
	return s.DirCache("object_id")
}

func (s directoryV1) FileInventoryListLog() string {
	return s.DirObjects("inventory_lists_log")
}

func (s directoryV1) DirInventoryLists() string {
	return s.DirObjects("inventory_lists")
}

func (s directoryV1) DirBlobs() string {
	return s.DirObjects("blobs")
}
