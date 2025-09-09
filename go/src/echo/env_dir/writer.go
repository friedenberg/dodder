package env_dir

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/markl_io"
)

// TODO fold into markl_io
type writer struct {
	repoolBufferedWriter  interfaces.FuncRepool
	digester              interfaces.BlobWriter
	tee                   io.Writer
	compressor, encrypter io.WriteCloser
	bufferedWriter        *bufio.Writer
}

func NewWriter(
	config Config,
	ioWriter io.Writer,
) (wrighter *writer, err error) {
	wrighter = &writer{}

	wrighter.bufferedWriter, wrighter.repoolBufferedWriter = pool.GetBufferedWriter(
		ioWriter,
	)

	if wrighter.encrypter, err = config.GetBlobEncryption().WrapWriter(
		wrighter.bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	wrighter.digester = markl_io.MakeWriter(config.hashType.Get(), nil)

	if wrighter.compressor, err = config.GetBlobCompression().WrapWriter(
		wrighter.encrypter,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	wrighter.tee = io.MultiWriter(wrighter.digester, wrighter.compressor)

	return
}

func (writer *writer) ReadFrom(r io.Reader) (n int64, err error) {
	if n, err = io.Copy(writer.tee, r); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (writer *writer) Write(p []byte) (n int, err error) {
	return writer.tee.Write(p)
}

func (writer *writer) WriteString(s string) (n int, err error) {
	return io.WriteString(writer.tee, s)
}

func (writer *writer) Close() (err error) {
	if err = writer.compressor.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = writer.encrypter.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = writer.bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	writer.bufferedWriter = nil
	writer.repoolBufferedWriter()

	return
}

func (writer *writer) GetMarklId() interfaces.MarklId {
	return writer.digester.GetMarklId()
}
