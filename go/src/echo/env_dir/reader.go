package env_dir

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

type reader struct {
	digester  interfaces.WriteBlobIdGetter
	decrypter io.Reader
	expander  io.ReadCloser
	tee       io.Reader
}

func NewReader(config Config, readSeeker io.ReadSeeker) (r *reader, err error) {
	r = &reader{}

	if r.decrypter, err = config.GetBlobEncryption().WrapReader(readSeeker); err != nil {
		err = errors.Wrap(err)
		return
	}

	if r.expander, err = config.GetBlobCompression().WrapReader(r.decrypter); err != nil {
		// TODO remove this when compression / encryption issues are resolved
		if _, err = readSeeker.Seek(0, io.SeekStart); err != nil {
			err = errors.Wrap(err)
			return
		}

		if r.expander, err = compression_type.CompressionTypeNone.WrapReader(
			readSeeker,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	r.digester = config.envDigest.MakeWriteDigester()
	r.tee = io.TeeReader(r.expander, r.digester)

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

func (reader *reader) GetBlobId() interfaces.BlobId {
	return reader.digester.GetBlobId()
}
