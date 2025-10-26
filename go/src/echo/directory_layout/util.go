package directory_layout

import (
	"fmt"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
)

const FileNameBlobStoreConfig = "dodder-blob_store-config"

func GetBlobStoreConfigPaths(
	ctx interfaces.ActiveContext,
	directoryLayout BlobStore,
) []string {
	var configPaths []string

	{
		var err error

		if configPaths, err = files.DirNames(
			filepath.Join(directoryLayout.DirBlobStoreConfigs()),
		); err != nil {
			ctx.Cancel(err)
			return configPaths
		}
	}

	return configPaths
}

func GetDefaultBlobStoreConfigPath(
	directoryLayout BlobStore,
) string {
	return directoryLayout.DirBlobStoreConfigs(
		fmt.Sprintf("%d-default.%s", 0, FileNameBlobStoreConfig),
	)
}

func GetBlobStoreConfigPath(
	directoryLayout BlobStore,
	currentBlobStoreCount int,
	name string,
) (dir string, target string) {
	dir = directoryLayout.DirBlobStoreConfigs()

	target = fmt.Sprintf(
		"%d-%s.%s",
		currentBlobStoreCount,
		name,
		FileNameBlobStoreConfig,
	)

	return dir, target
}

func PathBlobStore(
	layout BlobStore,
	targets ...string,
) interfaces.DirectoryLayoutPath {
	return layout.MakePathBlobStore(targets...)
}

func DirBlobStore(
	layout BlobStore,
	targets ...string,
) string {
	return PathBlobStore(layout, targets...).String()
}
