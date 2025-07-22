package object_inventory_format

import (
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type v4 struct{}

func (f v4) FormatPersistentMetadata(
	w1 io.Writer,
	c FormatterContext,
	o Options,
) (n int64, err error) {
	bufferedWriter := pool.GetBufioWriter().Get()
	defer pool.GetBufioWriter().Put(bufferedWriter)

	bufferedWriter.Reset(w1)
	defer errors.DeferredFlusher(&err, bufferedWriter)

	m := c.GetMetadata()

	mh := sha.MakeWriter(sha.Env{}, nil)
	mw := io.MultiWriter(bufferedWriter, mh)
	var (
		n1 int
		n2 int64
	)

	if !m.Blob.IsNull() {
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			mw,
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
			mw,
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

	for _, e := range quiter.SortedValues(es) {
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			mw,
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
		bufferedWriter,
		keyGattung.String(),
		c.GetObjectId().GetGenre().GetGenreString(),
	)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n1, err = ohio.WriteKeySpaceValueNewlineString(
		bufferedWriter,
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
			mw,
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
			mw,
			keyTai.String(),
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
			mw,
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
				bufferedWriter,
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
					bufferedWriter,
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
					bufferedWriter,
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

	if !m.Mutter().IsNull() && !o.ExcludeMutter {
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			mw,
			keyShasMutterMetadataKennungMutter.String(),
			m.Mutter().String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if o.PrintFinalSha {
		actual := mh.GetDigest()
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
			bufferedWriter,
			keySha.String(),
			digests.Format(actual),
		)

		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
