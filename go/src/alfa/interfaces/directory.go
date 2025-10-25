package interfaces

type (
	EnvVars = map[string]string

	EnvVarsAdder interface {
		AddToEnvVars(EnvVars)
	}

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
		GetDirCwd() DirectoryLayoutBaseEnvVar
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

		DirIndex(p ...string) string
		DirCacheRemoteInventoryListLog() string
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

	Directory interface {
		RepoDirectoryLayout
		Delete(...string) error
	}
)

func DirectoryLayoutPathBlobStore(
	layout BlobStoreDirectoryLayout,
	targets ...string,
) DirectoryLayoutPath {
	return layout.MakePathBlobStore(targets...)
}

func DirectoryLayoutDirBlobStore(
	layout BlobStoreDirectoryLayout,
	targets ...string,
) string {
	return DirectoryLayoutPathBlobStore(layout, targets...).String()
}
