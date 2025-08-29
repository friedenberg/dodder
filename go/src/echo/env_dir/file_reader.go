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
) (readCloser interfaces.ReadCloseBlobIdGetter, err error) {
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
	if objectReader.ReadCloseBlobIdGetter, err = NewReader(
		config,
		objectReader.file,
	); err != nil {
		if _, err = objectReader.file.Seek(0, io.SeekStart); err != nil {
			err = errors.Wrap(err)
			return
		}

		config = MakeConfig(
			config.hashType,
			config.funcJoin,
			config.GetBlobCompression(),
			&age.Age{},
		)

		if objectReader.ReadCloseBlobIdGetter, err = NewReader(config, objectReader.file); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	readCloser = objectReader

	return
}

type objectReader struct {
	file *os.File
	interfaces.ReadCloseBlobIdGetter
}

func (r objectReader) String() string {
	return r.file.Name()
}

func (ar objectReader) Close() (err error) {
	if ar.file == nil {
		err = errors.ErrorWithStackf("nil file")
		return
	}

	if ar.ReadCloseBlobIdGetter == nil {
		err = errors.ErrorWithStackf("nil object reader")
		return
	}

	if err = ar.ReadCloseBlobIdGetter.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = files.Close(ar.file); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
