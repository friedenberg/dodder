package blob_store_configs

import (
	"strconv"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type ConfigNamed struct {
	Index         int
	BasePath      string
	Name          string
	NameWithIndex string
	Config        TypedConfig
}

func GetDefaultBasePath(
	directoryLayout interfaces.BlobStoreDirectoryLayout,
	index int,
) string {
	return interfaces.DirectoryLayoutDirBlobStore(
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
	directoryLayout interfaces.BlobStoreDirectoryLayout,
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
