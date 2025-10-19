package env_repo

import (
	"fmt"

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

func (layout directoryLayoutV2) MakeDirData(targets ...string) string {
	if layout.xdg.Data.ActualValue == "" {
		panic(fmt.Sprintf("empty xdg data dir: %#v", layout.xdg.Data))
	}

	return layout.xdg.Data.MakePath(targets...).String()
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

func (layout directoryLayoutV2) DirIndex(p ...string) string {
	return layout.MakeDirData(append([]string{"index"}, p...)...)
}

func (layout directoryLayoutV2) DirCacheRepo(p ...string) string {
	return layout.xdg.Cache.MakePath(
		append([]string{"index", "repo"}, p...)...,
	).String()
}

func (layout directoryLayoutV2) DirLostAndFound() string {
	return layout.MakeDirData("lost_and_found")
}

func (layout directoryLayoutV2) DirIndexObjects() string {
	return layout.DirIndex("objects")
}

func (layout directoryLayoutV2) DirIndexObjectPointers() string {
	return layout.DirIndex("object_pointers/Page")
}

func (layout directoryLayoutV2) DirCacheRemoteInventoryListLog() string {
	return layout.xdg.Cache.MakePath("inventory_list_logs").String()
}

func (layout directoryLayoutV2) DirObjectId() string {
	return layout.MakeDirData("object_ids")
}

func (layout directoryLayoutV2) FileCacheObjectId() string {
	return layout.DirIndex("object_id")
}

func (layout directoryLayoutV2) FileInventoryListLog() string {
	return interfaces.DirectoryLayoutDirBlobStore(layout, "inventory_lists_log")
}
