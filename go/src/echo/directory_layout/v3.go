package directory_layout

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type v3 struct {
	xdg interfaces.DirectoryLayoutXDG
}

var (
	_ Repo        = v3{}
	_ repoUninitialized = &v3{}
)

func (layout *v3) initialize(
	xdg interfaces.DirectoryLayoutXDG,
) (err error) {
	layout.xdg = xdg
	return err
}

func (layout v3) MakeDirData(targets ...string) string {
	if layout.xdg.GetDirData().GetBaseEnvVarValue() == "" {
		panic(fmt.Sprintf("empty xdg data dir: %#v", layout.xdg.GetDirData()))
	}

	return layout.xdg.GetDirData().MakePath(targets...).String()
}

func (layout v3) MakePathBlobStore(
	targets ...string,
) interfaces.DirectoryLayoutPath {
	return layout.xdg.GetDirData().MakePath(
		stringSliceJoin("blob_stores", targets)...)
}

func (layout v3) DirBlobStoreConfigs(p ...string) string {
	return layout.MakeDirData(stringSliceJoin("blob_stores-configs", p)...)
}

func (layout v3) FileCacheDormant() string {
	return layout.MakeDirData("dormant")
}

func (layout v3) FileTags() string {
	return layout.MakeDirData("tags")
}

func (layout v3) FileLock() string {
	return layout.xdg.GetDirState().MakePath("lock").String()
}

func (layout v3) FileConfigPermanent() string {
	return layout.MakeDirData("config-permanent")
}

func (layout v3) FileConfigMutable() string {
	return layout.MakeDirData("config-mutable")
}

func (layout v3) DirDataIndex(p ...string) string {
	return layout.MakeDirData(stringSliceJoin("index", p)...)
}

func (layout v3) DirCacheRepo(p ...string) string {
	return layout.xdg.GetDirCache().MakePath(
		stringSliceJoin("repo", p)...).String()
}

func (layout v3) DirLostAndFound() string {
	return layout.MakeDirData("lost_and_found")
}

func (layout v3) DirIndexObjects() string {
	return layout.DirDataIndex("objects")
}

func (layout v3) DirIndexObjectPointers() string {
	return layout.DirDataIndex("object_pointers/Page")
}

func (layout v3) DirCacheRemoteInventoryListsLog() string {
	return layout.xdg.GetDirCache().MakePath(
		"remote_inventory_lists_log",
	).String()
}

func (layout v3) DirObjectId() string {
	return layout.MakeDirData("object_ids")
}

func (layout v3) FileCacheObjectId() string {
	return layout.DirDataIndex("object_id")
}

func (layout v3) FileInventoryListLog() string {
	return layout.MakeDirData("inventory_lists_log")
}

func (layout v3) DirsGenesis() []string {
	return []string{
		layout.DirObjectId(),
		layout.DirDataIndex(),
		layout.DirLostAndFound(),
		layout.DirBlobStoreConfigs(),
	}
}

// Deprecated

// TODO deprecate and remove
func (layout v3) DirFirstBlobStoreInventoryLists() string {
	panic(errors.Err405MethodNotAllowed)
}

// TODO deprecate and remove
func (layout v3) DirFirstBlobStoreBlobs() string {
	panic(errors.Err405MethodNotAllowed)
}
