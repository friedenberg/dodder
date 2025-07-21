package sha

import (
	"crypto/sha256"
	"hash"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

// TODO-P4 remove
type (
	ReadCloser  = interfaces.ReadCloserDigester
	WriteCloser = interfaces.WriteCloserDigester
)

type readCloser struct {
	tee  io.Reader
	r    io.Reader
	w    io.Writer
	hash hash.Hash
}

func MakeReadCloser(reader io.Reader) (src readCloser) {
	switch rt := reader.(type) {
	case *readCloser:
		src = *rt

	case readCloser:
		src = rt

	default:
		src.hash = sha256.New()
		src.r = reader
	}

	src.setupTee()

	return
}

func MakeReadCloserTee(r io.Reader, w io.Writer) (src readCloser) {
	switch rt := r.(type) {
	case *readCloser:
		src = *rt
		src.w = w

	case readCloser:
		src = rt
		src.w = w

	default:
		src.hash = sha256.New()
		src.r = r
		src.w = w
	}

	src.setupTee()

	return
}

func (readCloser *readCloser) setupTee() {
	if readCloser.w == nil {
		readCloser.tee = io.TeeReader(readCloser.r, readCloser.hash)
	} else {
		readCloser.tee = io.TeeReader(readCloser.r, io.MultiWriter(readCloser.hash, readCloser.w))
	}
}

func (readCloser readCloser) Seek(
	offset int64,
	whence int,
) (actual int64, err error) {
	seeker, ok := readCloser.r.(io.Seeker)

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
	if c, ok := readCloser.r.(io.Closer); ok {
		defer errors.DeferredCloser(&err, c)
	}

	if c, ok := readCloser.w.(io.Closer); ok {
		defer errors.DeferredCloser(&err, c)
	}

	return
}

func (readCloser readCloser) GetDigest() interfaces.Digest {
	return FromHash(readCloser.hash)
}

type nopReadCloser struct {
	io.ReadCloser
}

func MakeNopReadCloser(rc io.ReadCloser) ReadCloser {
	return nopReadCloser{
		ReadCloser: rc,
	}
}

func (nopReadCloser) Seek(offset int64, whence int) (actual int64, err error) {
	err = errors.ErrorWithStackf("seeking not supported")
	return
}

func (readCloser nopReadCloser) WriteTo(w io.Writer) (n int64, err error) {
	return io.Copy(w, readCloser.ReadCloser)
}

func (readCloser nopReadCloser) GetDigest() interfaces.Digest {
	return GetPool().Get()
}
