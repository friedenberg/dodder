package directory_layout

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blob_store_id"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

type (
	BlobStore interface {
		GetLocationType() blob_store_id.LocationType
		DirBlobStoreConfigs(p ...string) string
		MakePathBlobStore(...string) interfaces.DirectoryLayoutPath
	}

	Repo interface {
		MakeDirData(p ...string) string

		DirDataIndex(p ...string) string
		DirCacheRemoteInventoryListsLog() string
		DirIndexObjectPointers() string
		DirIndexObjects() string

		DirCacheRepo(p ...string) string

		DirLostAndFound() string
		DirObjectId() string

		FileCacheDormant() string
		FileCacheObjectId() string
		FileConfigMutable() string
		FileLock() string
		FileTags() string
		FileInventoryListLog() string

		// TODO remove from DirectoryLayout and move to method on EnvRepo
		FileConfigPermanent() string

		DirsGenesis() []string
	}

	Mutable interface {
		Delete(...string) error
	}

	RepoMutable interface {
		Repo
		Mutable
	}
)

type (
	XDG = xdg.XDG

	uninitializedXDG interface {
		initialize(XDG) error
	}

	blobStoreUninitialized interface {
		BlobStore
		uninitializedXDG
	}

	repoUninitialized interface {
		Repo
		uninitializedXDG
	}
)

func MakeRepo(
	storeVersion store_version.Version,
	xdg XDG,
) (Repo, error) {
	var repo repoUninitialized

	if storeVersion.LessOrEqual(store_version.V10) {
		repo = &v1{}
	} else if storeVersion.LessOrEqual(store_version.V12) {
		repo = &v2{}
	} else {
		repo = &v3{}
	}

	if err := repo.initialize(xdg); err != nil {
		err = errors.Wrap(err)
		return nil, err
	}

	return repo, nil
}

func MakeBlobStore(
	storeVersion store_version.Version,
	xdg XDG,
) (BlobStore, error) {
	var blobStore blobStoreUninitialized

	if storeVersion.LessOrEqual(store_version.V10) {
		blobStore = &v1{}
	} else if storeVersion.LessOrEqual(store_version.V11) {
		blobStore = &v2{}
	} else {
		blobStore = &v3{}
	}

	if err := blobStore.initialize(xdg); err != nil {
		err = errors.Wrap(err)
		return nil, err
	}

	return blobStore, nil
}
