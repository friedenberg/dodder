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
	"code.linenisgreat.com/dodder/go/src/delta/key_strings"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type v5 struct{}

func (f v5) FormatPersistentMetadata(
	w1 io.Writer,
	c FormatterContext,
	o Options,
) (n int64, err error) {
	w := pool.GetBufioWriter().Get()
	defer pool.GetBufioWriter().Put(w)

	w.Reset(w1)
	defer errors.DeferredFlusher(&err, w)

	m := c.GetMetadata()

	var (
		n1 int
		n2 int64
	)

	if !m.Blob.IsNull() {
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			w,
			keyAkte.String(),
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
			keyBezeichnung.String(),
			line,
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	es := m.GetTags()

	for _, e := range quiter.SortedValues[ids.Tag](es) {
		if e.IsVirtual() {
			continue
		}

		n1, err = ohio.WriteKeySpaceValueNewlineString(
			w,
			keyEtikett.String(),
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
		keyGattung.String(),
		c.GetObjectId().GetGenre().GetGenreString(),
	)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n1, err = ohio.WriteKeySpaceValueNewlineString(
		w,
		keyKennung.String(),
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
			keyKomment.String(),
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
			key_strings.Tai.String(),
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
			keyTyp.String(),
			m.GetType().String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if o.Verzeichnisse {
		if m.Cache.Dormant.Bool() {
			n1, err = ohio.WriteKeySpaceValueNewlineString(
				w,
				keyVerzeichnisseArchiviert.String(),
				m.Cache.Dormant.String(),
			)
			n += int64(n1)

			if err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		if m.Cache.GetExpandedTags().Len() > 0 {
			k := keyVerzeichnisseEtikettExpanded.String()

			for _, e := range quiter.SortedValues[ids.Tag](
				m.Cache.GetExpandedTags(),
			) {
				n1, err = ohio.WriteKeySpaceValueNewlineString(
					w,
					k,
					e.String(),
				)
				n += int64(n1)

				if err != nil {
					err = errors.Wrap(err)
					return
				}
			}
		}

		if m.Cache.GetImplicitTags().Len() > 0 {
			k := keyVerzeichnisseEtikettImplicit.String()

			for _, e := range quiter.SortedValues[ids.Tag](
				m.Cache.GetImplicitTags(),
			) {
				n2, err = ohio.WriteKeySpaceValueNewline(
					w,
					k,
					e.Bytes(),
				)
				n += int64(n2)

				if err != nil {
					err = errors.Wrap(err)
					return
				}
			}
		}
	}

	n1, err = writeShaKeyIfNotNull(
		w,
		keyShasMutterMetadataKennungMutter,
		m.GetMotherDigest(),
	)

	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO-P1 set value

	// if !m.Verzeichnisse.Sha.IsNull() &&
	// 	!m.Verzeichnisse.Sha.EqualsSha(actual) {
	// 	err = errors.Errorf(
	// 		"expected %q but got %q -> %q",
	// 		m.Verzeichnisse.Sha,
	// 		actual,
	// 		sb.String(),
	// 	)
	// 	return
	// }

	n1, err = ohio.WriteKeySpaceValueNewlineString(
		w,
		key_strings.Sha.String(),
		m.GetDigest().String(),
	)

	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (f v5) ParsePersistentMetadata(
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
				panic(
					fmt.Sprintf(
						"failed to write val to valBuffer. N: %d, Err: %s",
						n,
						err,
					),
				)
			}
		}

		switch {
		case key.Equal(keyAkte.Bytes()):
			if err = m.Blob.SetHexBytes(valBuffer.Bytes()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keyBezeichnung.Bytes()):
			if err = m.Description.Set(val.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keyEtikett.Bytes()):
			e := ids.GetTagPool().Get()

			if err = e.Set(val.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

			if err = m.AddTagPtr(e); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keyGattung.Bytes()):
			if err = g.Set(val.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keyKennung.Bytes()):
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

		case key.Equal(key_strings.Tai.Bytes()):
			if err = m.Tai.Set(val.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keyTyp.Bytes()):
			if err = m.Type.Set(val.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keyVerzeichnisseArchiviert.Bytes()):
			if err = m.Cache.Dormant.Set(val.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keyVerzeichnisseEtikettImplicit.Bytes()):
			if !o.Verzeichnisse {
				err = errors.ErrorWithStackf(
					"format specifies not to include Verzeichnisse but found %q",
					key,
				)
				return
			}

			e := ids.GetTagPool().Get()

			if err = e.Set(val.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

			if err = m.Cache.AddTagsImplicitPtr(e); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keyVerzeichnisseEtikettExpanded.Bytes()):
			if !o.Verzeichnisse {
				err = errors.ErrorWithStackf(
					"format specifies not to include Verzeichnisse but found %q",
					key,
				)
				return
			}

			e := ids.GetTagPool().Get()

			if err = e.Set(val.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

			if err = m.Cache.AddTagExpandedPtr(e); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keyShasMutterMetadataKennungMutter.Bytes()):
			if err = m.GetMotherDigest().SetHexBytes(val.Bytes()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(key_strings.Sha.Bytes()):
			if err = m.GetDigest().SetHexBytes(val.Bytes()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case key.Equal(keyKomment.Bytes()):
			m.Comments = append(m.Comments, val.String())

		default:
			err = errV4InvalidKey
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
