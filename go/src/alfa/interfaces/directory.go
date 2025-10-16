package interfaces

type (
	DirectoryLayoutPath interface {
		Stringer

		GetBaseEnvVar() DirectoryLayoutBaseEnvVar
		GetTarget() string

		GetTemplate() string
	}

	DirectoryLayoutBaseEnvVar interface {
		Stringer

		GetBaseEnvVarName() string
		GetBaseEnvVarValue() string

		MakePath(...string) DirectoryLayoutPath
	}

	DirectoryLayout interface {
		GetDirHome() DirectoryLayoutBaseEnvVar
		GetDirData() DirectoryLayoutBaseEnvVar
		GetDirConfig() DirectoryLayoutBaseEnvVar
		GetDirState() DirectoryLayoutBaseEnvVar
		GetDirCache() DirectoryLayoutBaseEnvVar
		GetDirRuntime() DirectoryLayoutBaseEnvVar
	}

	BlobStoreDirectoryLayout interface {
		DirFirstBlobStoreBlobs() string
		DirBlobStoreConfigs(p ...string) string

		MakePathBlobStore(...string) DirectoryLayoutPath
	}

	RepoDirectoryLayout interface {
		BlobStoreDirectoryLayout

		MakeDirData(p ...string) string

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

	Directory interface {
		RepoDirectoryLayout
		Delete(...string) error
	}
)

func DirectoryLayoutDirBlobStore(
	layout BlobStoreDirectoryLayout,
	targets ...string,
) string {
	return layout.MakePathBlobStore(targets...).String()
}
