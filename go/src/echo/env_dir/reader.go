package env_dir

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl_io"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

// TODO fold into markl_io
type reader struct {
	digester  interfaces.BlobWriter
	decrypter io.Reader
	expander  io.ReadCloser
	tee       io.Reader
}

func NewReader(
	config Config,
	readSeeker io.ReadSeeker,
) (reeder *reader, err error) {
	reeder = &reader{}

	if reeder.decrypter, err = config.GetBlobEncryption().WrapReader(
		readSeeker,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if reeder.expander, err = config.GetBlobCompression().WrapReader(
		reeder.decrypter,
	); err != nil {
		// TODO remove this when compression / encryption issues are resolved
		if _, err = readSeeker.Seek(0, io.SeekStart); err != nil {
			err = errors.Wrap(err)
			return
		}

		if reeder.expander, err = compression_type.CompressionTypeNone.WrapReader(
			readSeeker,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	reeder.digester = markl_io.MakeWriter(config.hashType.Get(), nil)
	reeder.tee = io.TeeReader(reeder.expander, reeder.digester)

	return
}

func (reader *reader) Seek(offset int64, whence int) (actual int64, err error) {
	seeker, ok := reader.decrypter.(io.Seeker)

	if !ok {
		err = errors.ErrorWithStackf("seeking not supported")
		return
	}

	return seeker.Seek(offset, whence)
}

func (reader *reader) WriteTo(w io.Writer) (n int64, err error) {
	return io.Copy(w, reader.tee)
}

func (reader *reader) Read(p []byte) (n int, err error) {
	return reader.tee.Read(p)
}

func (reader *reader) Close() (err error) {
	if err = reader.expander.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (reader *reader) GetMarklId() interfaces.MarklId {
	return reader.digester.GetMarklId()
}
