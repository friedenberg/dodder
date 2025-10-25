package directory_layout

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

// TODO examine the directories and use XDG more appropriately for them
type V3 struct {
	xdg xdg.XDG
}

var _ interfaces.RepoDirectoryLayout = V3{}

func (layout *V3) Initialize(
	xdg xdg.XDG,
) (err error) {
	layout.xdg = xdg
	return err
}

func (layout V3) MakeDirData(targets ...string) string {
	if layout.xdg.Data.ActualValue == "" {
		panic(fmt.Sprintf("empty xdg data dir: %#v", layout.xdg.Data))
	}

	return layout.xdg.Data.MakePath(targets...).String()
}

func (layout V3) MakePathBlobStore(
	targets ...string,
) interfaces.DirectoryLayoutPath {
	return layout.xdg.Data.MakePath(stringSliceJoin("blob_stores", targets)...)
}

func (layout V3) DirBlobStoreConfigs(p ...string) string {
	return layout.MakeDirData(append([]string{"blob_stores-configs"}, p...)...)
}

// TODO deprecate and remove
func (layout V3) DirFirstBlobStoreInventoryLists() string {
	return interfaces.DirectoryLayoutDirBlobStore(layout, "0")
}

// TODO deprecate and remove
func (layout V3) DirFirstBlobStoreBlobs() string {
	return interfaces.DirectoryLayoutDirBlobStore(layout, "0")
}

func (layout V3) FileCacheDormant() string {
	return layout.MakeDirData("dormant")
}

func (layout V3) FileTags() string {
	return layout.MakeDirData("tags")
}

func (layout V3) FileLock() string {
	return layout.xdg.State.MakePath("lock").String()
}

func (layout V3) FileConfigPermanent() string {
	return layout.MakeDirData("config-permanent")
}

func (layout V3) FileConfigMutable() string {
	return layout.MakeDirData("config-mutable")
}

func (layout V3) DirIndex(p ...string) string {
	return layout.MakeDirData(append([]string{"index"}, p...)...)
}

func (layout V3) DirCacheRepo(p ...string) string {
	return layout.xdg.Cache.MakePath(
		append([]string{"index", "repo"}, p...)...,
	).String()
}

func (layout V3) DirLostAndFound() string {
	return layout.MakeDirData("lost_and_found")
}

func (layout V3) DirIndexObjects() string {
	return layout.DirIndex("objects")
}

func (layout V3) DirIndexObjectPointers() string {
	return layout.DirIndex("object_pointers/Page")
}

func (layout V3) DirCacheRemoteInventoryListLog() string {
	return layout.xdg.Cache.MakePath("inventory_list_logs").String()
}

func (layout V3) DirObjectId() string {
	return layout.MakeDirData("object_ids")
}

func (layout V3) FileCacheObjectId() string {
	return layout.DirIndex("object_id")
}

func (layout V3) FileInventoryListLog() string {
	return layout.MakeDirData("inventory_lists_log")
}

func (layout V3) DirsGenesis() []string {
	return []string{
		layout.DirObjectId(),
		layout.DirIndex(),
		layout.DirLostAndFound(),
		layout.DirBlobStoreConfigs(),
	}
}
