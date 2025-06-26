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
	v *StoreVersion,
	p string,
) (err error) {
	if err = ReadFromFileOrVersion(v, p, store_version.VCurrent); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func ReadFromFileOrVersion(
	v *StoreVersion,
	p string,
	alternative StoreVersion,
) (err error) {
	var b []byte

	var f *os.File

	if f, err = files.Open(p); err != nil {
		if errors.IsNotExist(err) {
			*v = alternative
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	if b, err = io.ReadAll(f); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = v.Set(string(b)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
