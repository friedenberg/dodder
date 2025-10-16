package env_repo

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

// TODO examine the directories and use XDG more appropriately for them
type directoryLayoutV2 struct {
	xdg xdg.XDG
}

var _ interfaces.RepoDirectoryLayout = directoryLayoutV2{}

func (layout *directoryLayoutV2) initDirectoryLayout(
	xdg xdg.XDG,
) (err error) {
	layout.xdg = xdg
	return err
}

func (layout directoryLayoutV2) MakeDirData(p ...string) string {
	return layout.xdg.Data.MakePath(p...).String()
}

func (layout directoryLayoutV2) MakePathBlobStore(
	targets ...string,
) interfaces.DirectoryLayoutPath {
	return layout.xdg.Data.MakePath(stringSliceJoin("blob_stores", targets)...)
}

func (layout directoryLayoutV2) DirBlobStoreConfigs(p ...string) string {
	return layout.MakeDirData(append([]string{"blob_stores-configs"}, p...)...)
}

// TODO deprecate and remove
func (layout directoryLayoutV2) DirFirstBlobStoreInventoryLists() string {
	return interfaces.DirectoryLayoutDirBlobStore(layout, "0")
}

// TODO deprecate and remove
func (layout directoryLayoutV2) DirFirstBlobStoreBlobs() string {
	return interfaces.DirectoryLayoutDirBlobStore(layout, "0")
}

func (layout directoryLayoutV2) FileCacheDormant() string {
	return layout.MakeDirData("dormant")
}

func (layout directoryLayoutV2) FileTags() string {
	return layout.MakeDirData("tags")
}

func (layout directoryLayoutV2) FileLock() string {
	return layout.xdg.State.MakePath("lock").String()
}

func (layout directoryLayoutV2) FileConfigPermanent() string {
	return layout.MakeDirData("config-permanent")
}

func (layout directoryLayoutV2) FileConfigMutable() string {
	return layout.MakeDirData("config-mutable")
}

func (layout directoryLayoutV2) DirCache(p ...string) string {
	return layout.MakeDirData(append([]string{"index"}, p...)...)
}

func (layout directoryLayoutV2) DirCacheRepo(p ...string) string {
	// TODO switch to XDG cache
	// return filepath.Join(stringSliceJoin(s.Cache, "repo", p...)...)
	return layout.MakeDirData(append([]string{"index", "repo"}, p...)...)
}

func (layout directoryLayoutV2) DirLostAndFound() string {
	return layout.MakeDirData("lost_and_found")
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
	return layout.MakeDirData("object_ids")
}

func (layout directoryLayoutV2) FileCacheObjectId() string {
	return layout.DirCache("object_id")
}

func (layout directoryLayoutV2) FileInventoryListLog() string {
	return interfaces.DirectoryLayoutDirBlobStore(layout, "inventory_lists_log")
}
