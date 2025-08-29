package merkle_ids

import (
	"hash"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type readCloser struct {
	envDigest interfaces.EnvBlobId
	tee       io.Reader
	reader    io.Reader
	writer    io.Writer
	hash      hash.Hash
}

func MakeReadCloser(
	envDigest interfaces.EnvBlobId,
	reader io.Reader,
) (src readCloser) {
	src.envDigest = envDigest

	switch rt := reader.(type) {
	case *readCloser:
		src = *rt

	case readCloser:
		src = rt

	default:
		src.hash, _ = src.envDigest.GetHash()
		src.reader = reader
	}

	src.setupTee()

	return
}

func MakeReadCloserTee(
	envDigest interfaces.EnvBlobId,
	reader io.Reader,
	writer io.Writer,
) (src readCloser) {
	src.envDigest = envDigest

	switch rt := reader.(type) {
	case *readCloser:
		src = *rt
		src.writer = writer

	case readCloser:
		src = rt
		src.writer = writer

	default:
		src.hash, _ = src.envDigest.GetHash()
		src.reader = reader
		src.writer = writer
	}

	src.setupTee()

	return
}

func (readCloser *readCloser) setupTee() {
	if readCloser.writer == nil {
		readCloser.tee = io.TeeReader(
			readCloser.reader,
			readCloser.hash,
		)
	} else {
		readCloser.tee = io.TeeReader(
			readCloser.reader,
			io.MultiWriter(readCloser.hash, readCloser.writer),
		)
	}
}

func (readCloser readCloser) Seek(
	offset int64,
	whence int,
) (actual int64, err error) {
	seeker, ok := readCloser.reader.(io.Seeker)

	if !ok {
		err = errors.ErrorWithStackf("seeking not supported")
		return
	}

	return seeker.Seek(offset, whence)
}

func (readCloser readCloser) WriteTo(w io.Writer) (n int64, err error) {
	// TODO-P3 determine why something in the copy returns an EOF
	if n, err = io.Copy(w, readCloser.tee); err != nil {
		if errors.IsEOF(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (readCloser readCloser) Read(b []byte) (n int, err error) {
	return readCloser.tee.Read(b)
}

func (readCloser readCloser) Close() (err error) {
	if c, ok := readCloser.reader.(io.Closer); ok {
		defer errors.DeferredCloser(&err, c)
	}

	if c, ok := readCloser.writer.(io.Closer); ok {
		defer errors.DeferredCloser(&err, c)
	}

	return
}

func (readCloser readCloser) GetBlobId() interfaces.BlobId {
	digest, err := readCloser.envDigest.MakeDigestFromHash(readCloser.hash)
	errors.PanicIfError(err)
	return digest
}

type nopReadCloser struct {
	envDigest interfaces.EnvBlobId
	io.ReadCloser
}

func MakeNopReadCloser(
	hash interfaces.EnvBlobId,
	readCloser io.ReadCloser,
) interfaces.ReadCloseBlobIdGetter {
	return nopReadCloser{
		envDigest:  hash,
		ReadCloser: readCloser,
	}
}

func (nopReadCloser) Seek(offset int64, whence int) (actual int64, err error) {
	err = errors.ErrorWithStackf("seeking not supported")
	return
}

func (readCloser nopReadCloser) WriteTo(writer io.Writer) (n int64, err error) {
	return io.Copy(writer, readCloser.ReadCloser)
}

func (readCloser nopReadCloser) GetBlobId() interfaces.BlobId {
	return readCloser.envDigest.GetBlobId()
}
