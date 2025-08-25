package merkle_ids

import (
	"hash"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var poolWriter = pool.MakePool[writer](nil, nil)

func MakeWriterWithRepool(
	envDigest interfaces.EnvBlobId,
	in io.Writer,
) (writer *writer, repool func()) {
	writer = poolWriter.Get()
	writer.Reset(envDigest, in)

	repool = func() {
		PutWriter(writer)
	}

	return
}

func MakeWriter(
	envDigest interfaces.EnvBlobId,
	in io.Writer,
) (writer *writer) {
	writer, _ = MakeWriterWithRepool(envDigest, in)
	return
}

func PutWriter(writer *writer) {
	poolWriter.Put(writer)
}

type writer struct {
	envDigest interfaces.EnvBlobId
	closed    bool
	in        io.Writer
	closer    io.Closer
	writer    io.Writer
	hash      hash.Hash
}

func (writer *writer) Reset(envDigest interfaces.EnvBlobId, in io.Writer) {
	writer.envDigest = envDigest

	if in == nil {
		in = io.Discard
	}

	writer.in = in

	if closer, ok := in.(io.Closer); ok {
		writer.closer = closer
	}

	if writer.hash == nil {
		writer.hash, _ = writer.envDigest.GetHash()
	} else {
		writer.hash.Reset()
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

func (writer *writer) GetBlobId() interfaces.BlobId {
	digest, err := writer.envDigest.MakeDigestFromHash(writer.hash)
	errors.PanicIfError(err)

	return digest
}
