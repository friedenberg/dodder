package object_inventory_format

import (
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/blob_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/key_strings"
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

	metadata := c.GetMetadata()

	digester, repool := blob_ids.MakeWriterWithRepool(sha.Env{}, nil)
	defer repool()

	multiWriter := io.MultiWriter(bufferedWriter, digester)

	var (
		n1 int
		n2 int64
	)

	if !metadata.Blob.IsNull() {
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			multiWriter,
			keyAkte.String(),
			metadata.Blob.String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	lines := strings.Split(metadata.Description.String(), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		n1, err = ohio.WriteKeySpaceValueNewlineString(
			multiWriter,
			keyBezeichnung.String(),
			line,
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	es := metadata.GetTags()

	for _, e := range quiter.SortedValues(es) {
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			multiWriter,
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

	for _, k := range metadata.Comments {
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			multiWriter,
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
			multiWriter,
			key_strings.Tai.String(),
			metadata.Tai.String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if !metadata.Type.IsEmpty() {
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			multiWriter,
			keyTyp.String(),
			metadata.GetType().String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if o.Verzeichnisse {
		if metadata.Cache.Dormant.Bool() {
			n1, err = ohio.WriteKeySpaceValueNewlineString(
				bufferedWriter,
				keyVerzeichnisseArchiviert.String(),
				metadata.Cache.Dormant.String(),
			)
			n += int64(n1)

			if err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		if metadata.Cache.GetExpandedTags().Len() > 0 {
			k := keyVerzeichnisseEtikettExpanded.String()

			for _, e := range quiter.SortedValues[ids.Tag](
				metadata.Cache.GetExpandedTags(),
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

		if metadata.Cache.GetImplicitTags().Len() > 0 {
			k := keyVerzeichnisseEtikettImplicit.String()

			for _, e := range quiter.SortedValues[ids.Tag](
				metadata.Cache.GetImplicitTags(),
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

	if !metadata.GetMotherDigest().IsNull() && !o.ExcludeMutter {
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			multiWriter,
			keyShasMutterMetadataKennungMutter.String(),
			metadata.GetMotherDigest().String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if o.PrintFinalSha {
		actual := digester.GetBlobId()
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
			key_strings.Sha.String(),
			blob_ids.Format(actual),
		)

		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
