package directory_layout

import (
	"fmt"
	"path/filepath"
	"strconv"

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

func GetDefaultBlobStore(
	directoryLayout BlobStore,
) BlobStorePath {
	return GetBlobStorePath(
		directoryLayout,
		0,
		"default",
	)
}

func GetBlobStorePath(
	directoryLayout BlobStore,
	currentBlobStoreCount int,
	name string,
) BlobStorePath {
	dirConfig := directoryLayout.DirBlobStoreConfigs()

	targetConfig := fmt.Sprintf(
		"%d-%s.%s",
		currentBlobStoreCount,
		name,
		FileNameBlobStoreConfig,
	)

	return BlobStorePath{
		Base:   DirBlobStore(directoryLayout, strconv.Itoa(0)),
		Config: filepath.Join(dirConfig, targetConfig),
	}
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
