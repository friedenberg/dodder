package markl_io

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type readCloser struct {
	tee    io.Reader
	reader io.Reader
	writer io.Writer
	hash   interfaces.Hash
}

var _ interfaces.BlobReader = readCloser{}

func MakeReadCloser(
	hash interfaces.Hash,
	reader io.Reader,
) (src readCloser) {
	switch rt := reader.(type) {
	case *readCloser:
		src = *rt

	case readCloser:
		src = rt

	default:
		src.hash = hash
		src.reader = reader
	}

	src.setupTee()

	return
}

func MakeReadCloserTee(
	hash interfaces.Hash,
	reader io.Reader,
	writer io.Writer,
) (src readCloser) {
	switch rt := reader.(type) {
	case *readCloser:
		src = *rt
		src.writer = writer

	case readCloser:
		src = rt
		src.writer = writer

	default:
		src.hash = hash
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

func (readCloser readCloser) GetMarklId() interfaces.MarklId {
	digest, _ := readCloser.hash.GetMarklId()
	return digest
}

type nopReadCloser struct {
	hash interfaces.Hash
	io.ReadCloser
}

var _ interfaces.BlobReader = nopReadCloser{}

func MakeNopReadCloser(
	hash interfaces.Hash,
	readCloser io.ReadCloser,
) interfaces.BlobReader {
	return nopReadCloser{
		hash:       hash,
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

func (readCloser nopReadCloser) GetMarklId() interfaces.MarklId {
	id, _ := readCloser.hash.GetMarklId()
	return id
}
