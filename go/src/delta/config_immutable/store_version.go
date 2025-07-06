package config_immutable

import (
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
)

type StoreVersion = store_version.Version

func ReadFromFile(
	version *StoreVersion,
	path string,
) (err error) {
	if err = ReadFromFileOrVersion(version, path, store_version.VCurrent); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func ReadFromFileOrVersion(
	version *StoreVersion,
	path string,
	alternative StoreVersion,
) (err error) {
	var bytes []byte

	var file *os.File

	if file, err = files.Open(path); err != nil {
		if errors.IsNotExist(err) {
			*version = alternative
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	if bytes, err = io.ReadAll(file); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = version.Set(string(bytes)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
