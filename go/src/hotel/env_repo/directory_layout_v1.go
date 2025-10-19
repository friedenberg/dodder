package env_repo

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

type directoryLayoutV1 struct {
	xdg.XDG
}

var _ interfaces.RepoDirectoryLayout = directoryLayoutV1{}

func (layout *directoryLayoutV1) initDirectoryLayout(
	xdg xdg.XDG,
) (err error) {
	layout.XDG = xdg
	return err
}

func (layout directoryLayoutV1) MakePathBlobStore(
	targets ...string,
) interfaces.DirectoryLayoutPath {
	return layout.XDG.Data.MakePath(stringSliceJoin("blob_stores", targets)...)
}

func (layout directoryLayoutV1) FileCacheDormant() string {
	return layout.MakeDirData("dormant")
}

func (layout directoryLayoutV1) FileTags() string {
	return layout.MakeDirData("tags")
}

func (layout directoryLayoutV1) FileLock() string {
	return layout.State.MakePath("lock").String()
}

func (layout directoryLayoutV1) FileConfigPermanent() string {
	return layout.MakeDirData("config-permanent")
}

func (layout directoryLayoutV1) FileConfigMutable() string {
	return layout.MakeDirData("config-mutable")
}

func (layout directoryLayoutV1) MakeDirData(p ...string) string {
	return layout.Data.MakePath(p...).String()
}

func (layout directoryLayoutV1) DirIndex(p ...string) string {
	return layout.MakeDirData(append([]string{"cache"}, p...)...)
}

func (layout directoryLayoutV1) DirCacheRepo(p ...string) string {
	// TODO switch to XDG cache
	// return filepath.Join(stringSliceJoin(s.Cache, "repo", p...)...)
	return layout.MakeDirData(append([]string{"cache", "repo"}, p...)...)
}

func (layout directoryLayoutV1) DirBlobStoreConfigs(p ...string) string {
	return interfaces.DirectoryLayoutDirBlobStore(layout, p...)
}

func (layout directoryLayoutV1) DirLostAndFound() string {
	return layout.MakeDirData("lost_and_found")
}

func (layout directoryLayoutV1) DirIndexObjects() string {
	return layout.DirIndex("objects")
}

func (layout directoryLayoutV1) DirIndexObjectPointers() string {
	return layout.DirIndex("object_pointers")
}

func (layout directoryLayoutV1) DirCacheRemoteInventoryListLog() string {
	return layout.DirIndex("inventory_list_logs")
}

func (layout directoryLayoutV1) DirObjectId() string {
	return layout.MakeDirData("object_ids")
}

func (layout directoryLayoutV1) FileCacheObjectId() string {
	return layout.DirIndex("object_id")
}

func (layout directoryLayoutV1) FileInventoryListLog() string {
	return interfaces.DirectoryLayoutDirBlobStore(layout, "inventory_lists_log")
}

func (layout directoryLayoutV1) DirFirstBlobStoreInventoryLists() string {
	return interfaces.DirectoryLayoutDirBlobStore(layout, "inventory_lists")
}

func (layout directoryLayoutV1) DirFirstBlobStoreBlobs() string {
	return interfaces.DirectoryLayoutDirBlobStore(layout, "blobs")
}
