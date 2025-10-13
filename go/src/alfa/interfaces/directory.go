package interfaces

type BlobStoreDirectoryLayout interface {
	DirFirstBlobStoreBlobs() string
	DirBlobStoreConfigs(p ...string) string
	DirBlobStores(p ...string) string
}

type RepoDirectoryLayout interface {
	BlobStoreDirectoryLayout

	Dir(p ...string) string
	DirDodder(p ...string) string

	DirFirstBlobStoreInventoryLists() string

	// TODO rename Cache to Index
	DirCache(p ...string) string
	DirCacheInventoryListLog() string
	DirCacheObjectPointers() string
	DirCacheObjects() string
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
}

type Directory interface {
	RepoDirectoryLayout
	Delete(...string) error
}
