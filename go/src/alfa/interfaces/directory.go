package interfaces

type DirectoryPaths interface {
	Dir(p ...string) string
	DirCache(p ...string) string
	DirCacheObjectPointers() string
	DirCacheObjects() string
	DirCacheInventoryListLog() string
	DirCacheRepo(p ...string) string
	DirLostAndFound() string
	DirObjectId() string
	DirObjects(p ...string) string
	DirInventoryLists() string
	DirBlobs() string
	DirZit(p ...string) string
	FileCacheDormant() string
	FileCacheObjectId() string
	FileConfigMutable() string
	FileConfigPermanent() string
	FileLock() string
	FileTags() string
	FileInventoryListLog() string
}

type Directory interface {
	DirectoryPaths
	Delete(...string) error
}

type CacheIOFactory interface {
	ReadCloserCache(string) (ShaReadCloser, error)
	WriteCloserCache(string) (ShaWriteCloser, error)
}
