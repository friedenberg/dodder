package directory_layout

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

// TODO examine the directories and use XDG more appropriately for them
type v2 struct {
	xdg interfaces.DirectoryLayoutXDG
}

var (
	_ Repo              = v2{}
	_ repoUninitialized = &v2{}
)

func (layout *v2) initialize(
	xdg interfaces.DirectoryLayoutXDG,
) (err error) {
	layout.xdg = xdg
	return err
}

func (layout v2) MakeDirData(targets ...string) string {
	if layout.xdg.GetDirData().GetBaseEnvVarValue() == "" {
		panic(fmt.Sprintf("empty xdg data dir: %#v", layout.xdg.GetDirData()))
	}

	return layout.xdg.GetDirData().MakePath(targets...).String()
}

func (layout v2) MakePathBlobStore(
	targets ...string,
) interfaces.DirectoryLayoutPath {
	return layout.xdg.GetDirData().MakePath(
		stringSliceJoin("blob_stores", targets)...)
}

func (layout v2) DirBlobStoreConfigs(p ...string) string {
	return layout.MakeDirData(append([]string{"blob_stores-configs"}, p...)...)
}

func (layout v2) FileCacheDormant() string {
	return layout.MakeDirData("dormant")
}

func (layout v2) FileTags() string {
	return layout.MakeDirData("tags")
}

func (layout v2) FileLock() string {
	return layout.xdg.GetDirState().MakePath("lock").String()
}

func (layout v2) FileConfigPermanent() string {
	return layout.MakeDirData("config-permanent")
}

func (layout v2) FileConfigMutable() string {
	return layout.MakeDirData("config-mutable")
}

func (layout v2) DirDataIndex(p ...string) string {
	return layout.MakeDirData(append([]string{"index"}, p...)...)
}

func (layout v2) DirCacheRepo(p ...string) string {
	return layout.xdg.GetDirCache().MakePath(
		append([]string{"index", "repo"}, p...)...,
	).String()
}

func (layout v2) DirLostAndFound() string {
	return layout.MakeDirData("lost_and_found")
}

func (layout v2) DirIndexObjects() string {
	return layout.DirDataIndex("objects")
}

func (layout v2) DirIndexObjectPointers() string {
	return layout.DirDataIndex("object_pointers/Page")
}

func (layout v2) DirCacheRemoteInventoryListsLog() string {
	return layout.xdg.GetDirCache().MakePath("inventory_list_logs").String()
}

func (layout v2) DirObjectId() string {
	return layout.MakeDirData("object_ids")
}

func (layout v2) FileCacheObjectId() string {
	return layout.DirDataIndex("object_id")
}

func (layout v2) FileInventoryListLog() string {
	return DirBlobStore(
		layout,
		"inventory_lists_log",
	)
}

func (layout v2) DirsGenesis() []string {
	return []string{
		layout.DirObjectId(),
		layout.DirDataIndex(),
		layout.DirLostAndFound(),
		layout.DirBlobStoreConfigs(),
		DirBlobStore(layout, "0"),
	}
}
