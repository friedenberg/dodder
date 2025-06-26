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
	ReadCloser  = interfaces.ShaReadCloser
	WriteCloser = interfaces.ShaWriteCloser
)

type readCloser struct {
	tee  io.Reader
	r    io.Reader
	w    io.Writer
	hash hash.Hash
}

func MakeReadCloser(r io.Reader) (src readCloser) {
	switch rt := r.(type) {
	case *readCloser:
		src = *rt

	case readCloser:
		src = rt

	default:
		src.hash = sha256.New()
		src.r = r
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

func (src *readCloser) setupTee() {
	if src.w == nil {
		src.tee = io.TeeReader(src.r, src.hash)
	} else {
		src.tee = io.TeeReader(src.r, io.MultiWriter(src.hash, src.w))
	}
}

func (r readCloser) Seek(offset int64, whence int) (actual int64, err error) {
	seeker, ok := r.r.(io.Seeker)

	if !ok {
		err = errors.ErrorWithStackf("seeking not supported")
		return
	}

	return seeker.Seek(offset, whence)
}

func (r readCloser) WriteTo(w io.Writer) (n int64, err error) {
	// TODO-P3 determine why something in the copy returns an EOF
	if n, err = io.Copy(w, r.tee); err != nil {
		if errors.IsEOF(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (r readCloser) Read(b []byte) (n int, err error) {
	return r.tee.Read(b)
}

func (r readCloser) Close() (err error) {
	if c, ok := r.r.(io.Closer); ok {
		defer errors.DeferredCloser(&err, c)
	}

	if c, ok := r.w.(io.Closer); ok {
		defer errors.DeferredCloser(&err, c)
	}

	return
}

func (r readCloser) GetShaLike() interfaces.Sha {
	return FromHash(r.hash)
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

func (nrc nopReadCloser) WriteTo(w io.Writer) (n int64, err error) {
	return io.Copy(w, nrc.ReadCloser)
}

func (nrc nopReadCloser) GetShaLike() interfaces.Sha {
	return &Sha{}
}
