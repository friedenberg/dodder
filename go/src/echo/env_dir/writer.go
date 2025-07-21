package env_dir

import (
	"bufio"
	"crypto/sha256"
	"hash"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

type Writer interface {
	interfaces.WriteCloseDigester
	interfaces.Digester
}

type writer struct {
	hash            hash.Hash
	tee             io.Writer
	wCompress, wAge io.WriteCloser
	wBuf            *bufio.Writer
}

func NewWriter(
	config Config,
	ioWriter io.Writer,
) (w *writer, err error) {
	w = &writer{}

	w.wBuf = bufio.NewWriter(ioWriter)

	if w.wAge, err = config.GetBlobEncryption().WrapWriter(w.wBuf); err != nil {
		err = errors.Wrap(err)
		return
	}

	w.hash = sha256.New()

	if w.wCompress, err = config.GetBlobCompression().WrapWriter(w.wAge); err != nil {
		err = errors.Wrap(err)
		return
	}

	w.tee = io.MultiWriter(w.hash, w.wCompress)

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
	if err = w.wCompress.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = w.wAge.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = w.wBuf.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (w *writer) GetDigest() interfaces.Digest {
	return sha.FromHash(w.hash)
}
