package directory_layout

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

// TODO examine the directories and use XDG more appropriately for them
type V2 struct {
	xdg xdg.XDG
}

var _ interfaces.RepoDirectoryLayout = V2{}

func (layout *V2) Initialize(
	xdg xdg.XDG,
) (err error) {
	layout.xdg = xdg
	return err
}

func (layout V2) MakeDirData(targets ...string) string {
	if layout.xdg.Data.ActualValue == "" {
		panic(fmt.Sprintf("empty xdg data dir: %#v", layout.xdg.Data))
	}

	return layout.xdg.Data.MakePath(targets...).String()
}

func (layout V2) MakePathBlobStore(
	targets ...string,
) interfaces.DirectoryLayoutPath {
	return layout.xdg.Data.MakePath(stringSliceJoin("blob_stores", targets)...)
}

func (layout V2) DirBlobStoreConfigs(p ...string) string {
	return layout.MakeDirData(append([]string{"blob_stores-configs"}, p...)...)
}

// TODO deprecate and remove
func (layout V2) DirFirstBlobStoreInventoryLists() string {
	return DirBlobStore(layout, "0")
}

// TODO deprecate and remove
func (layout V2) DirFirstBlobStoreBlobs() string {
	return DirBlobStore(layout, "0")
}

func (layout V2) FileCacheDormant() string {
	return layout.MakeDirData("dormant")
}

func (layout V2) FileTags() string {
	return layout.MakeDirData("tags")
}

func (layout V2) FileLock() string {
	return layout.xdg.State.MakePath("lock").String()
}

func (layout V2) FileConfigPermanent() string {
	return layout.MakeDirData("config-permanent")
}

func (layout V2) FileConfigMutable() string {
	return layout.MakeDirData("config-mutable")
}

func (layout V2) DirIndex(p ...string) string {
	return layout.MakeDirData(append([]string{"index"}, p...)...)
}

func (layout V2) DirCacheRepo(p ...string) string {
	return layout.xdg.Cache.MakePath(
		append([]string{"index", "repo"}, p...)...,
	).String()
}

func (layout V2) DirLostAndFound() string {
	return layout.MakeDirData("lost_and_found")
}

func (layout V2) DirIndexObjects() string {
	return layout.DirIndex("objects")
}

func (layout V2) DirIndexObjectPointers() string {
	return layout.DirIndex("object_pointers/Page")
}

func (layout V2) DirCacheRemoteInventoryListLog() string {
	return layout.xdg.Cache.MakePath("inventory_list_logs").String()
}

func (layout V2) DirObjectId() string {
	return layout.MakeDirData("object_ids")
}

func (layout V2) FileCacheObjectId() string {
	return layout.DirIndex("object_id")
}

func (layout V2) FileInventoryListLog() string {
	return DirBlobStore(
		layout,
		"inventory_lists_log",
	)
}

func (layout V2) DirsGenesis() []string {
	return []string{
		layout.DirObjectId(),
		layout.DirIndex(),
		layout.DirLostAndFound(),
		layout.DirBlobStoreConfigs(),
		layout.DirFirstBlobStoreInventoryLists(),
		layout.DirFirstBlobStoreBlobs(),
		DirBlobStore(layout, "0"),
	}
}
