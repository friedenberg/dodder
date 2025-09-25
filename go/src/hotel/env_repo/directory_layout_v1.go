package env_repo

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

type directoryLayoutV1 struct {
	xdg.XDG
}

func (layout *directoryLayoutV1) initDirectoryLayout(
	xdg xdg.XDG,
) (err error) {
	layout.XDG = xdg
	return err
}

func (layout directoryLayoutV1) GetDirectoryPaths() interfaces.DirectoryLayout {
	return layout
}

func (layout directoryLayoutV1) FileCacheDormant() string {
	return layout.DirDodder("dormant")
}

func (layout directoryLayoutV1) FileTags() string {
	return layout.DirDodder("tags")
}

func (layout directoryLayoutV1) FileLock() string {
	return filepath.Join(layout.State, "lock")
}

func (layout directoryLayoutV1) FileConfigPermanent() string {
	return layout.DirDodder("config-permanent")
}

func (layout directoryLayoutV1) FileConfigMutable() string {
	return layout.DirDodder("config-mutable")
}

func (layout directoryLayoutV1) Dir(p ...string) string {
	return filepath.Join(stringSliceJoin(layout.Data, p)...)
}

func (layout directoryLayoutV1) DirDodder(p ...string) string {
	return layout.Dir(p...)
}

func (layout directoryLayoutV1) DirCache(p ...string) string {
	return layout.DirDodder(append([]string{"cache"}, p...)...)
}

func (layout directoryLayoutV1) DirCacheRepo(p ...string) string {
	// TODO switch to XDG cache
	// return filepath.Join(stringSliceJoin(s.Cache, "repo", p...)...)
	return layout.DirDodder(append([]string{"cache", "repo"}, p...)...)
}

func (layout directoryLayoutV1) DirBlobStores(p ...string) string {
	return layout.DirDodder(append([]string{"objects"}, p...)...)
}

func (layout directoryLayoutV1) DirBlobStoreConfigs(p ...string) string {
	return layout.DirBlobStores(p...)
}

func (layout directoryLayoutV1) DirLostAndFound() string {
	return layout.DirDodder("lost_and_found")
}

func (layout directoryLayoutV1) DirCacheObjects() string {
	return layout.DirCache("objects")
}

func (layout directoryLayoutV1) DirCacheObjectPointers() string {
	return layout.DirCache("object_pointers")
}

func (layout directoryLayoutV1) DirCacheInventoryListLog() string {
	return layout.DirCache("inventory_list_logs")
}

func (layout directoryLayoutV1) DirObjectId() string {
	return layout.DirDodder("object_ids")
}

func (layout directoryLayoutV1) FileCacheObjectId() string {
	return layout.DirCache("object_id")
}

func (layout directoryLayoutV1) FileInventoryListLog() string {
	return layout.DirBlobStores("inventory_lists_log")
}

func (layout directoryLayoutV1) DirFirstBlobStoreInventoryLists() string {
	return layout.DirBlobStores("inventory_lists")
}

func (layout directoryLayoutV1) DirFirstBlobStoreBlobs() string {
	return layout.DirBlobStores("blobs")
}
