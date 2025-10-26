package directory_layout

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

type V1 struct {
	xdg.XDG
}

var _ interfaces.RepoDirectoryLayout = V1{}

func (layout *V1) Initialize(
	xdg xdg.XDG,
) (err error) {
	layout.XDG = xdg
	return err
}

func (layout V1) MakePathBlobStore(
	targets ...string,
) interfaces.DirectoryLayoutPath {
	return layout.XDG.Data.MakePath(stringSliceJoin("blob_stores", targets)...)
}

func (layout V1) FileCacheDormant() string {
	return layout.MakeDirData("dormant")
}

func (layout V1) FileTags() string {
	return layout.MakeDirData("tags")
}

func (layout V1) FileLock() string {
	return layout.State.MakePath("lock").String()
}

func (layout V1) FileConfigPermanent() string {
	return layout.MakeDirData("config-permanent")
}

func (layout V1) FileConfigMutable() string {
	return layout.MakeDirData("config-mutable")
}

func (layout V1) MakeDirData(p ...string) string {
	return layout.Data.MakePath(p...).String()
}

func (layout V1) DirIndex(p ...string) string {
	return layout.MakeDirData(append([]string{"cache"}, p...)...)
}

func (layout V1) DirCacheRepo(p ...string) string {
	// TODO switch to XDG cache
	// return filepath.Join(stringSliceJoin(s.Cache, "repo", p...)...)
	return layout.MakeDirData(append([]string{"cache", "repo"}, p...)...)
}

func (layout V1) DirBlobStoreConfigs(p ...string) string {
	return DirBlobStore(layout, p...)
}

func (layout V1) DirLostAndFound() string {
	return layout.MakeDirData("lost_and_found")
}

func (layout V1) DirIndexObjects() string {
	return layout.DirIndex("objects")
}

func (layout V1) DirIndexObjectPointers() string {
	return layout.DirIndex("object_pointers")
}

func (layout V1) DirCacheRemoteInventoryListLog() string {
	return layout.DirIndex("inventory_list_logs")
}

func (layout V1) DirObjectId() string {
	return layout.MakeDirData("object_ids")
}

func (layout V1) FileCacheObjectId() string {
	return layout.DirIndex("object_id")
}

func (layout V1) FileInventoryListLog() string {
	return DirBlobStore(
		layout,
		"inventory_lists_log",
	)
}

func (layout V1) DirFirstBlobStoreInventoryLists() string {
	return DirBlobStore(
		layout,
		"inventory_lists",
	)
}

func (layout V1) DirFirstBlobStoreBlobs() string {
	return DirBlobStore(layout, "blobs")
}

func (layout V1) DirsGenesis() []string {
	return []string{
		layout.DirObjectId(),
		layout.DirIndex(),
		layout.DirLostAndFound(),
		layout.DirBlobStoreConfigs(),
		layout.DirFirstBlobStoreInventoryLists(),
		DirBlobStore(layout, "0"),
	}
}
