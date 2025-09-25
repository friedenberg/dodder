package env_dir

import (
	"bytes"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl_io"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

// TODO fold into markl_io
type blobReader struct {
	readSeeker io.ReadSeeker
	digester   interfaces.BlobWriter
	decrypter  io.Reader
	expander   io.ReadCloser
	tee        io.Reader
}

func NewReader(
	config Config,
	readSeeker io.ReadSeeker,
) (reader *blobReader, err error) {
	reader = &blobReader{
		readSeeker: readSeeker,
	}

	if reader.decrypter, err = config.GetBlobEncryption().WrapReader(
		reader.readSeeker,
	); err != nil {
		err = errors.Wrap(err)
		return reader, err
	}

	if reader.expander, err = config.GetBlobCompression().WrapReader(
		reader.decrypter,
	); err != nil {
		// TODO remove this when compression / encryption issues are resolved
		if _, err = reader.readSeeker.Seek(0, io.SeekStart); err != nil {
			err = errors.Wrap(err)
			return reader, err
		}

		if reader.expander, err = compression_type.CompressionTypeNone.WrapReader(
			reader.readSeeker,
		); err != nil {
			err = errors.Wrap(err)
			return reader, err
		}
	}

	reader.digester = markl_io.MakeWriter(config.hashFormat.GetHash(), nil)
	reader.tee = io.TeeReader(reader.expander, reader.digester)

	return reader, err
}

// TODO fold into markl_io
func NewFileReader(
	config Config,
	path string,
) (blobReader interfaces.BlobReader, err error) {
	var readSeeker io.ReadSeeker

	if path == "-" {
		readSeeker = os.Stdin
	} else {
		var file *os.File

		if file, err = files.Open(path); err != nil {
			if errors.IsNotExist(err) {
				readSeeker = bytes.NewReader(nil)
				err = nil
			} else {
				err = errors.Wrap(err)
				return blobReader, err
			}
		} else {
			readSeeker = file
		}
	}

	// try the existing options. if they fail, try without encryption
	if blobReader, err = NewReader(
		config,
		readSeeker,
	); err != nil {
		if _, err = readSeeker.Seek(0, io.SeekStart); err != nil {
			err = errors.Wrap(err)
			return blobReader, err
		}

		config = MakeConfig(
			config.hashFormat,
			config.funcJoin,
			config.GetBlobCompression(),
			nil,
		)

		if blobReader, err = NewReader(
			config,
			readSeeker,
		); err != nil {
			err = errors.Wrap(err)
			return blobReader, err
		}
	}

	return blobReader, err
}

func (reader *blobReader) Seek(
	offset int64,
	whence int,
) (actual int64, err error) {
	seeker, ok := reader.decrypter.(io.Seeker)

	if !ok {
		err = errors.ErrorWithStackf("seeking not supported")
		return actual, err
	}

	return seeker.Seek(offset, whence)
}

func (reader *blobReader) ReadAt(p []byte, off int64) (n int, err error) {
	readerAt, ok := reader.decrypter.(io.ReaderAt)

	if !ok {
		err = errors.ErrorWithStackf("reading at not supported")
		return n, err
	}

	return readerAt.ReadAt(p, off)
}

func (reader *blobReader) WriteTo(w io.Writer) (n int64, err error) {
	return io.Copy(w, reader.tee)
}

func (reader *blobReader) Read(p []byte) (n int, err error) {
	return reader.tee.Read(p)
}

func (reader *blobReader) Close() (err error) {
	if err = reader.expander.Close(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if closer, ok := reader.readSeeker.(io.Closer); ok {
		if err = closer.Close(); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (reader *blobReader) GetMarklId() interfaces.MarklId {
	return reader.digester.GetMarklId()
}
