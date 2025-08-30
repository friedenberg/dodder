package env_dir

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
)

type Writer interface {
	interfaces.WriteCloseBlobIdGetter
	interfaces.BlobIdGetter
}

type writer struct {
	digester              interfaces.WriteBlobIdGetter
	tee                   io.Writer
	compressor, encrypter io.WriteCloser
	buffer                *bufio.Writer
}

func NewWriter(
	config Config,
	ioWriter io.Writer,
) (w *writer, err error) {
	w = &writer{}

	// TODO use pool
	w.buffer = bufio.NewWriter(ioWriter)

	if w.encrypter, err = config.GetBlobEncryption().WrapWriter(w.buffer); err != nil {
		err = errors.Wrap(err)
		return
	}

	w.digester = markl.MakeWriter(config.hashType.Get(), nil)

	if w.compressor, err = config.GetBlobCompression().WrapWriter(w.encrypter); err != nil {
		err = errors.Wrap(err)
		return
	}

	w.tee = io.MultiWriter(w.digester, w.compressor)

	return
}

func (w *writer) ReadFrom(r io.Reader) (n int64, err error) {
	if n, err = io.Copy(w.tee, r); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (w *writer) Write(p []byte) (n int, err error) {
	return w.tee.Write(p)
}

func (w *writer) WriteString(s string) (n int, err error) {
	return io.WriteString(w.tee, s)
}

func (w *writer) Close() (err error) {
	if err = w.compressor.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = w.encrypter.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = w.buffer.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (w *writer) GetBlobId() interfaces.BlobId {
	return w.digester.GetBlobId()
}
