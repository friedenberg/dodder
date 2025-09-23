package env_dir

import (
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
)

// TODO fold into markl_io
func NewFileReader(
	config Config,
	path string,
) (readCloser interfaces.BlobReader, err error) {
	objectReader := objectReader{}

	if path == "-" {
		objectReader.file = os.Stdin
	} else {
		if objectReader.file, err = files.Open(path); err != nil {
			err = errors.Wrap(err)
			return readCloser, err
		}
	}

	// try the existing options. if they fail, try without encryption
	if objectReader.BlobReader, err = NewReader(
		config,
		objectReader.file,
	); err != nil {
		if _, err = objectReader.file.Seek(0, io.SeekStart); err != nil {
			err = errors.Wrap(err)
			return readCloser, err
		}

		config = MakeConfig(
			config.hashFormat,
			config.funcJoin,
			config.GetBlobCompression(),
			nil,
		)

		if objectReader.BlobReader, err = NewReader(config, objectReader.file); err != nil {
			err = errors.Wrap(err)
			return readCloser, err
		}
	}

	readCloser = objectReader

	return readCloser, err
}

type objectReader struct {
	file *os.File
	interfaces.BlobReader
}

func (reader objectReader) String() string {
	return reader.file.Name()
}

func (reader objectReader) Close() (err error) {
	if reader.file == nil {
		err = errors.ErrorWithStackf("nil file")
		return err
	}

	if reader.BlobReader == nil {
		err = errors.ErrorWithStackf("nil object reader")
		return err
	}

	if err = reader.BlobReader.Close(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = files.Close(reader.file); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
