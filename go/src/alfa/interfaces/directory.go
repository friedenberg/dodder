package interfaces

type DirectoryLayout interface {
	Dir(p ...string) string
	DirDodder(p ...string) string

	DirFirstBlobStoreBlobs() string
	DirFirstBlobStoreInventoryLists() string
	DirBlobStoreConfigs(p ...string) string
	DirBlobStores(p ...string) string

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
	DirectoryLayout
	Delete(...string) error
}

type CacheIOFactory interface {
	ReadCloserCache(string) (ReadCloseMarklIdGetter, error)
	WriteCloserCache(string) (WriteCloseMarklIdGetter, error)
}
