package env_repo

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

// TODO examine the directories and use XDG more appropriately for them
type directoryLayoutV2 struct {
	xdg xdg.XDG
}

func (layout *directoryLayoutV2) initDirectoryLayout(
	xdg xdg.XDG,
) (err error) {
	layout.xdg = xdg
	return err
}

func (layout directoryLayoutV2) GetDirectoryPaths() interfaces.RepoDirectoryLayout {
	return layout
}

func (layout directoryLayoutV2) Dir(p ...string) string {
	return filepath.Join(stringSliceJoin(layout.xdg.Data, p)...)
}

func (layout directoryLayoutV2) DirDodder(p ...string) string {
	return layout.Dir(p...)
}

func (layout directoryLayoutV2) DirBlobStores(p ...string) string {
	return layout.DirDodder(append([]string{"blob_stores"}, p...)...)
}

func (layout directoryLayoutV2) DirBlobStoreConfigs(p ...string) string {
	return layout.DirDodder(append([]string{"blob_stores-configs"}, p...)...)
}

// TODO deprecate and remove
func (layout directoryLayoutV2) DirFirstBlobStoreInventoryLists() string {
	return layout.DirBlobStores("0")
}

// TODO deprecate and remove
func (layout directoryLayoutV2) DirFirstBlobStoreBlobs() string {
	return layout.DirBlobStores("0")
}

func (layout directoryLayoutV2) FileCacheDormant() string {
	return layout.DirDodder("dormant")
}

func (layout directoryLayoutV2) FileTags() string {
	return layout.DirDodder("tags")
}

func (layout directoryLayoutV2) FileLock() string {
	return filepath.Join(layout.xdg.State, "lock")
}

func (layout directoryLayoutV2) FileConfigPermanent() string {
	return layout.DirDodder("config-permanent")
}

func (layout directoryLayoutV2) FileConfigMutable() string {
	return layout.DirDodder("config-mutable")
}

func (layout directoryLayoutV2) DirCache(p ...string) string {
	return layout.DirDodder(append([]string{"index"}, p...)...)
}

func (layout directoryLayoutV2) DirCacheRepo(p ...string) string {
	// TODO switch to XDG cache
	// return filepath.Join(stringSliceJoin(s.Cache, "repo", p...)...)
	return layout.DirDodder(append([]string{"index", "repo"}, p...)...)
}

func (layout directoryLayoutV2) DirLostAndFound() string {
	return layout.DirDodder("lost_and_found")
}

func (layout directoryLayoutV2) DirCacheObjects() string {
	return layout.DirCache("objects")
}

func (layout directoryLayoutV2) DirCacheObjectPointers() string {
	return layout.DirCache("object_pointers/Page")
}

func (layout directoryLayoutV2) DirCacheInventoryListLog() string {
	return layout.DirCache("inventory_list_logs")
}

func (layout directoryLayoutV2) DirObjectId() string {
	return layout.DirDodder("object_ids")
}

func (layout directoryLayoutV2) FileCacheObjectId() string {
	return layout.DirCache("object_id")
}

func (layout directoryLayoutV2) FileInventoryListLog() string {
	return layout.DirBlobStores("inventory_lists_log")
}
