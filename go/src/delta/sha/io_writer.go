package sha

import (
	"hash"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type writer struct {
	closed bool
	in     io.Writer
	c      io.Closer
	w      io.Writer
	h      hash.Hash
	sh     Sha
}

func MakeWriter(in io.Writer) (writer *writer) {
	writer = GetPoolWriter().Get()
	writer.Reset(in)

	return
}

func (w *writer) ReadFrom(r io.Reader) (n int64, err error) {
	if n, err = io.Copy(w.w, r); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (w *writer) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func (w *writer) WriteString(v string) (n int, err error) {
	if stringWriter, ok := w.w.(io.StringWriter); ok {
		return stringWriter.WriteString(v)
	} else {
		return io.WriteString(w.w, v)
	}
}

func (w *writer) Close() (err error) {
	w.closed = true
	w.setShaLikeIfNecessary()

	if w.c == nil {
		return
	}

	if err = w.c.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (w *writer) setShaLikeIfNecessary() {
	if w.h != nil {
		errors.PanicIfError(w.sh.SetFromHash(w.h))

		poolHash256.Put(w.h)
		w.h = nil
	}
}

func (w *writer) GetDigest() (digest interfaces.Digest) {
	w.setShaLikeIfNecessary()
	digest = &w.sh

	return
}

func (writer *writer) Reset(in io.Writer) {
	if in == nil {
		in = io.Discard
	}

	writer.in = in

	if c, ok := in.(io.Closer); ok {
		writer.c = c
	}

	if writer.h == nil {
		writer.h = poolHash256.Get()
	} else {
		writer.h.Reset()
	}

	writer.w = io.MultiWriter(writer.h, writer.in)
}
