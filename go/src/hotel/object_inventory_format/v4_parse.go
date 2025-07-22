package object_inventory_format

import (
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type key struct {
	bytes        []byte
	includeInSha bool
}

func (f v4) ParsePersistentMetadata(
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

		case key.Equal(keyTai.Bytes()):
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

		case key.Equal(keySha.Bytes()):
			if err = m.GetSha().SetHexBytes(val.Bytes()); err != nil {
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

	// if m.Verzeichnisse.Sha.IsNull() {
	// } else if !m.Verzeichnisse.Sha.EqualsSha(actual) &&
	// o.IncludeVerzeichnisse {
	// 	err = errors.Errorf(
	// 		"expected %q but got %q",
	// 		m.Verzeichnisse.Sha,
	// 		actual,
	// 	)
	// 	return
	// }

	return
}
