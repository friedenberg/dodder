package markl

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var poolWriter = pool.Make[writer](nil, nil)

func MakeWriterWithRepool(
	hash interfaces.Hash,
	in io.Writer,
) (writer *writer, repool func()) {
	writer = poolWriter.Get()
	writer.Reset(hash, in)

	repool = func() {
		PutWriter(writer)
	}

	return
}

func MakeWriter(
	hash interfaces.Hash,
	in io.Writer,
) (writer *writer) {
	writer, _ = MakeWriterWithRepool(hash, in)
	return
}

func PutWriter(writer *writer) {
	poolWriter.Put(writer)
}

type writer struct {
	closed bool
	in     io.Writer
	closer io.Closer
	writer io.Writer
	hash   interfaces.Hash
}

var _ interfaces.MarklIdGetter = &writer{}

func (writer *writer) Reset(hash interfaces.Hash, in io.Writer) {
	if writer.hash == nil || writer.hash.GetType() != hash.GetType() {
		writer.hash = hash
	} else {
		writer.hash.Reset()
	}

	if in == nil {
		in = io.Discard
	}

	writer.in = in

	if closer, ok := in.(io.Closer); ok {
		writer.closer = closer
	}

	writer.writer = io.MultiWriter(writer.hash, writer.in)
}

func (writer *writer) ReadFrom(r io.Reader) (n int64, err error) {
	if n, err = io.Copy(writer.writer, r); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (writer *writer) Write(p []byte) (n int, err error) {
	return writer.writer.Write(p)
}

func (writer *writer) WriteString(v string) (n int, err error) {
	if stringWriter, ok := writer.writer.(io.StringWriter); ok {
		return stringWriter.WriteString(v)
	} else {
		return io.WriteString(writer.writer, v)
	}
}

func (writer *writer) Close() (err error) {
	writer.closed = true

	if writer.closer == nil {
		return
	}

	if err = writer.closer.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (writer *writer) GetMarklId() interfaces.MarklId {
	digest, _ := writer.hash.GetMarklId()
	return digest
}
