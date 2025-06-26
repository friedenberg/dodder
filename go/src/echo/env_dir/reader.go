package env_dir

import (
	"crypto/sha256"
	"hash"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

type reader struct {
	hash      hash.Hash
	decrypter io.Reader
	expander  io.ReadCloser
	tee       io.Reader
}

func NewReader(options ReadOptions) (r *reader, err error) {
	r = &reader{}

	if r.decrypter, err = options.GetBlobEncryption().WrapReader(options.File); err != nil {
		err = errors.Wrap(err)
		return
	}

	if r.expander, err = options.GetBlobCompression().WrapReader(r.decrypter); err != nil {
		// TODO remove this when compression / encryption issues are resolved
		if _, err = options.File.Seek(0, io.SeekStart); err != nil {
			err = errors.Wrap(err)
			return
		}

		if r.expander, err = config_immutable.CompressionTypeNone.WrapReader(
			options.File,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	r.hash = sha256.New()
	r.tee = io.TeeReader(r.expander, r.hash)

	return
}

func (r *reader) Seek(offset int64, whence int) (actual int64, err error) {
	seeker, ok := r.decrypter.(io.Seeker)

	if !ok {
		err = errors.ErrorWithStackf("seeking not supported")
		return
	}

	return seeker.Seek(offset, whence)
}

func (r *reader) WriteTo(w io.Writer) (n int64, err error) {
	return io.Copy(w, r.tee)
}

func (r *reader) Read(p []byte) (n int, err error) {
	return r.tee.Read(p)
}

func (r *reader) Close() (err error) {
	if err = r.expander.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (r *reader) GetShaLike() (s interfaces.Sha) {
	s = sha.FromHash(r.hash)

	return
}
