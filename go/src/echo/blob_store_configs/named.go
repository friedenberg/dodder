package blob_store_configs

import (
	"strconv"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/directory_layout"
)

type ConfigNamed struct {
	Path   directory_layout.BlobStorePath
	Config TypedConfig
}

func GetDefaultBasePath(
	directoryLayout directory_layout.BlobStore,
	index int,
) string {
	return directory_layout.DirBlobStore(
		directoryLayout,
		strconv.Itoa(index),
	)
}

func GetBasePath(
	config Config,
) (string, bool) {
	configLocal, ok := config.(configLocal)

	if !ok {
		return "", false
	}

	basePath := configLocal.getBasePath()
	return basePath, basePath != ""
}

func GetBasePathOrDefault(
	config ConfigNamed,
	directoryLayout directory_layout.BlobStore,
	index int,
) (string, bool) {
	configLocal, ok := config.Config.Blob.(configLocal)

	if !ok {
		return "", false
	}

	basePath := configLocal.getBasePath()

	if basePath != "" {
		return basePath, true
	}

	return GetDefaultBasePath(directoryLayout, index), true
}

func SetBasePath(
	config Config,
	basePath interfaces.DirectoryLayoutPath,
) (err error) {
	configLocalMutable, ok := config.(configLocalMutable)

	if !ok {
		err = errors.Errorf("setting base path not supported on %T", config)
		return err
	}

	configLocalMutable.setBasePath(basePath.String())

	return err
}
