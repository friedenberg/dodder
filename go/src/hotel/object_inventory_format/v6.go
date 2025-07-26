package object_inventory_format

import (
	"fmt"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/keys"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type v6 struct{}

func (f v6) FormatPersistentMetadata(
	w1 io.Writer,
	c FormatterContext,
	o Options,
) (n int64, err error) {
	w := pool.GetBufioWriter().Get()
	defer pool.GetBufioWriter().Put(w)

	w.Reset(w1)
	defer errors.DeferredFlusher(&err, w)

	m := c.GetMetadata()

	var n1 int

	if !m.Blob.IsNull() {
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			w,
			keys.KeyBlob.String(),
			m.Blob.String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	lines := strings.Split(m.Description.String(), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		n1, err = ohio.WriteKeySpaceValueNewlineString(
			w,
			keys.KeyDescription.String(),
			line,
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	es := m.GetTags()

	for _, e := range quiter.SortedValues(es) {
		if e.IsVirtual() {
			continue
		}

		n1, err = ohio.WriteKeySpaceValueNewlineString(
			w,
			keys.KeyTag.String(),
			e.String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	n1, err = ohio.WriteKeySpaceValueNewlineString(
		w,
		keys.KeyGenre.String(),
		c.GetObjectId().GetGenre().GetGenreString(),
	)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n1, err = ohio.WriteKeySpaceValueNewlineString(
		w,
		keys.KeyObjectId.String(),
		c.GetObjectId().String(),
	)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, k := range m.Comments {
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			w,
			keys.KeyComment.String(),
			k,
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if o.Tai {
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			w,
			keys.KeyTai.String(),
			m.Tai.String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if !m.Type.IsEmpty() {
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			w,
			keys.KeyType.String(),
			m.GetType().String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	n1, err = ohio.WriteKeySpaceValueNewlineString(
		w,
		keys.KeySha.String(),
		m.GetDigest().String(),
	)

	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (f v6) ParsePersistentMetadata(
	r *catgut.RingBuffer,
	c ParserContext,
	o Options,
) (n int64, err error) {
	m := c.GetMetadata()

	var (
		g genres.Genre
		k *ids.ObjectId
	)

	var (
		valBuffer      catgut.String
		line, key, val catgut.Slice
		ok             bool
	)

	lineNo := 0

	for {
		line, err = r.PeekUpto('\n')

		if errors.IsNotNilAndNotEOF(err) {
			break
		}

		if line.Len() == 0 {
			break
		}

		key, val, ok = line.Cut(' ')

		if !ok {
			err = makeErrWithBytes(ErrV4ExpectedSpaceSeparatedKey, line.Bytes())
			break
		}

		if key.Len() == 0 {
			err = makeErrWithBytes(errV4EmptyKey, line.Bytes())
			break
		}

		{
			valBuffer.Reset()
			n, err := val.WriteTo(&valBuffer)

			if n != int64(val.Len()) || err != nil {
				panic(fmt.Sprintf("failed to write val to valBuffer. N: %d, Err: %s", n, err))
			}
		}

		switch {
		case key.Equal(keys.KeyBlob.Bytes()):
			if err = m.Blob.SetHexBytes(valBuffer.Bytes()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keys.KeyDescription.Bytes()):
			if err = m.Description.Set(val.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keys.KeyTag.Bytes()):
			e := ids.GetTagPool().Get()

			if err = e.Set(val.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

			if err = m.AddTagPtr(e); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keys.KeyGenre.Bytes()):
			if err = g.Set(val.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keys.KeyObjectId.Bytes()):
			k = ids.GetObjectIdPool().Get()
			defer ids.GetObjectIdPool().Put(k)

			if err = k.SetWithGenre(val.String(), g); err != nil {
				err = errors.Wrap(err)
				return
			}

			if err = c.SetObjectIdLike(k); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keys.KeyTai.Bytes()):
			if err = m.Tai.Set(val.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keys.KeyType.Bytes()):
			if err = m.Type.Set(val.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keys.KeySha.Bytes()):
			if err = m.GetDigest().SetHexBytes(val.Bytes()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keys.KeyComment.Bytes()):
			m.Comments = append(m.Comments, val.String())

		default:
			err = errV6InvalidKey
		}

		// Key Space Value Newline
		thisN := int64(key.Len() + 1 + val.Len() + 1)
		n += thisN

		lineNo++

		r.AdvanceRead(int(thisN))
	}

	if n == 0 {
		if err == nil {
			err = io.EOF
		}

		return
	}

	return
}
