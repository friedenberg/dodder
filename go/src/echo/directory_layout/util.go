package directory_layout

import (
	"fmt"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

const FileNameBlobStoreConfig = "dodder-blob_store-config"

func GetBlobStoreConfigPaths(
	ctx interfaces.ActiveContext,
	directoryLayout BlobStore,
) []string {
	var configPaths []string

	{
		var err error

		if configPaths, err = filepath.Glob(
			DirBlobStore(directoryLayout, fmt.Sprintf("*/%s", FileNameBlobStoreConfig)),
		); err != nil {
			ctx.Cancel(err)
			return configPaths
		}
	}

	return configPaths
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
