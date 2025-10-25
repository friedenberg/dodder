package directory_layout

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
)

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
