package directory_layout

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blob_store_id"
)

type v1 struct {
	XDG
}

var (
	_ Repo              = v1{}
	_ repoUninitialized = &v1{}
)

func (layout *v1) initialize(
	xdg XDG,
) (err error) {
	layout.XDG = xdg
	return err
}

func (layout v1) MakePathBlobStore(
	targets ...string,
) interfaces.DirectoryLayoutPath {
	return layout.GetDirData().MakePath(
		stringSliceJoin("blob_stores", targets)...)
}

func (layout v1) FileCacheDormant() string {
	return layout.MakeDirData("dormant")
}

func (layout v1) FileTags() string {
	return layout.MakeDirData("tags")
}

func (layout v1) FileLock() string {
	return layout.GetDirState().MakePath("lock").String()
}

func (layout v1) FileConfigPermanent() string {
	return layout.MakeDirData("config-permanent")
}

func (layout v1) FileConfigMutable() string {
	return layout.MakeDirData("config-mutable")
}

func (layout v1) MakeDirData(p ...string) string {
	return layout.GetDirData().MakePath(p...).String()
}

func (layout v1) DirDataIndex(p ...string) string {
	return layout.MakeDirData(append([]string{"cache"}, p...)...)
}

func (layout v1) DirCacheRepo(p ...string) string {
	// TODO switch to XDG cache
	// return filepath.Join(stringSliceJoin(s.Cache, "repo", p...)...)
	return layout.MakeDirData(append([]string{"cache", "repo"}, p...)...)
}

func (layout v1) DirBlobStoreConfigs(p ...string) string {
	return DirBlobStore(layout, p...)
}

func (layout v1) DirLostAndFound() string {
	return layout.MakeDirData("lost_and_found")
}

func (layout v1) DirIndexObjects() string {
	return layout.DirDataIndex("objects")
}

func (layout v1) DirIndexObjectPointers() string {
	return layout.DirDataIndex("object_pointers")
}

func (layout v1) DirCacheRemoteInventoryListsLog() string {
	return layout.DirDataIndex("inventory_list_logs")
}

func (layout v1) DirObjectId() string {
	return layout.MakeDirData("object_ids")
}

func (layout v1) FileCacheObjectId() string {
	return layout.DirDataIndex("object_id")
}

func (layout v1) FileInventoryListLog() string {
	return DirBlobStore(
		layout,
		"inventory_lists_log",
	)
}

func (layout v1) DirFirstBlobStoreInventoryLists() string {
	return DirBlobStore(
		layout,
		"inventory_lists",
	)
}

func (layout v1) DirFirstBlobStoreBlobs() string {
	return DirBlobStore(layout, "blobs")
}

func (layout v1) DirsGenesis() []string {
	return []string{
		layout.DirObjectId(),
		layout.DirDataIndex(),
		layout.DirLostAndFound(),
		layout.DirBlobStoreConfigs(),
		layout.DirFirstBlobStoreInventoryLists(),
		DirBlobStore(layout, "0"),
	}
}

func (layout v1) GetLocationType() blob_store_id.LocationType {
	return layout.XDG.GetLocationType()
}

func (layout v1) cloneUninitialized() uninitializedXDG {
	return &layout
}
