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
	directoryLayout interfaces.BlobStoreDirectoryLayout,
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
	directoryLayout interfaces.BlobStoreDirectoryLayout,
) string {
	return directoryLayout.DirBlobStoreConfigs(
		fmt.Sprintf("%d-default.%s", 0, FileNameBlobStoreConfig),
	)
}

func GetBlobStoreConfigPath(
	directoryLayout interfaces.BlobStoreDirectoryLayout,
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
