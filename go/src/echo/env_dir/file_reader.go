package env_dir

import (
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/age"
)

func NewFileReader(
	config Config,
	path string,
) (readCloser interfaces.ReadCloserDigester, err error) {
	objectReader := objectReader{}

	if path == "-" {
		objectReader.file = os.Stdin
	} else {
		if objectReader.file, err = files.Open(path); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	// try the existing options. if they fail, try without encryption
	if objectReader.ReadCloserDigester, err = NewReader(config, objectReader.file); err != nil {
		if _, err = objectReader.file.Seek(0, io.SeekStart); err != nil {
			err = errors.Wrap(err)
			return
		}

		config = MakeConfig(
			config.funcJoin,
			config.GetBlobCompression(),
			&age.Age{},
		)

		if objectReader.ReadCloserDigester, err = NewReader(config, objectReader.file); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	readCloser = objectReader

	return
}

type objectReader struct {
	file *os.File
	interfaces.ReadCloserDigester
}

func (r objectReader) String() string {
	return r.file.Name()
}

func (ar objectReader) Close() (err error) {
	if ar.file == nil {
		err = errors.ErrorWithStackf("nil file")
		return
	}

	if ar.ReadCloserDigester == nil {
		err = errors.ErrorWithStackf("nil object reader")
		return
	}

	if err = ar.ReadCloserDigester.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = files.Close(ar.file); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
