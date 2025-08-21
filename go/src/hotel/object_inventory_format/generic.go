package object_inventory_format

import (
	"fmt"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blob_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/key_strings"
	"code.linenisgreat.com/dodder/go/src/delta/key_strings_german"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

type (
	Metadata = object_metadata.Metadata
	Sha      = sha.Sha
)

const (
	KeyFormatV5Metadata               = "Metadatei"
	KeyFormatV5MetadataWithoutTai     = "MetadateiSansTai"
	KeyFormatV5MetadataObjectIdParent = "MetadateiKennungMutter"
	KeyFormatV6Metadata               = "Metadata"
	KeyFormatV6MetadataWithoutTai     = "MetadataWithoutTai"
	KeyFormatV6MetadataObjectIdParent = "MetadataObjectIdParent"
)

var (
	FormatKeys = []string{
		KeyFormatV5Metadata,
		KeyFormatV5MetadataWithoutTai,
		KeyFormatV5MetadataObjectIdParent,
		KeyFormatV6Metadata,
		KeyFormatV6MetadataWithoutTai,
		KeyFormatV6MetadataObjectIdParent,
	}

	FormatKeysV5 = []string{
		KeyFormatV5Metadata,
		KeyFormatV5MetadataWithoutTai,
		KeyFormatV5MetadataObjectIdParent,
	}

	FormatKeysV6 = []string{
		KeyFormatV6Metadata,
		KeyFormatV6MetadataWithoutTai,
		KeyFormatV6MetadataObjectIdParent,
	}

	// TODO remove local aliases and only use german_keys
	keyAkte                            = key_strings_german.Akte
	keyBezeichnung                     = key_strings_german.Bezeichnung
	keyEtikett                         = key_strings_german.Etikett
	keyGattung                         = key_strings_german.Gattung
	keyKennung                         = key_strings_german.Kennung
	keyKomment                         = key_strings_german.Komment
	keyTyp                             = key_strings_german.Typ
	keyShasMutterMetadataKennungMutter = key_strings_german.ShasMutterMetadateiKennungMutter
	keyVerzeichnisseArchiviert         = key_strings_german.VerzeichnisseArchiviert
	keyVerzeichnisseEtikettImplicit    = key_strings_german.VerzeichnisseEtikettImplicit
	keyVerzeichnisseEtikettExpanded    = key_strings_german.VerzeichnisseEtikettExpanded
)

type FormatGeneric struct {
	key  string
	keys []*catgut.String
}

type formats struct {
	metadataSansTai        FormatGeneric
	metadata               FormatGeneric
	metadataObjectIdParent FormatGeneric
}

func (fs formats) MetadataSansTai() FormatGeneric {
	return fs.metadataSansTai
}

func (fs formats) Metadata() FormatGeneric {
	return fs.metadata
}

func (fs formats) MetadataObjectIdParent() FormatGeneric {
	return fs.metadataObjectIdParent
}

var Formats formats

func init() {
	Formats.metadata.key = KeyFormatV5Metadata
	Formats.metadata.keys = []*catgut.String{
		keyAkte,
		keyBezeichnung,
		keyEtikett,
		keyTyp,
		key_strings.Tai,
	}

	Formats.metadataSansTai.key = KeyFormatV5MetadataWithoutTai
	Formats.metadataSansTai.keys = []*catgut.String{
		keyAkte,
		keyBezeichnung,
		keyEtikett,
		keyTyp,
	}

	Formats.metadataObjectIdParent.key = KeyFormatV5MetadataObjectIdParent
	Formats.metadataObjectIdParent.keys = []*catgut.String{
		keyAkte,
		keyBezeichnung,
		keyEtikett,
		keyKennung,
		keyTyp,
		key_strings.Tai,
		keyShasMutterMetadataKennungMutter,
	}
}

func FormatForKeyError(k string) (fo FormatGeneric, err error) {
	switch k {
	case KeyFormatV5Metadata:
		fo = Formats.metadata

	case KeyFormatV5MetadataWithoutTai:
		fo = Formats.metadataSansTai

	case KeyFormatV5MetadataObjectIdParent:
		fo = Formats.metadataObjectIdParent

	default:
		err = errInvalidGenericFormat(k)
		return
	}

	return
}

func FormatForKey(k string) FormatGeneric {
	f, err := FormatForKeyError(k)
	errors.PanicIfError(err)
	return f
}

func (f FormatGeneric) WriteMetadataTo(
	w io.Writer,
	c FormatterContext,
) (n int64, err error) {
	var n1 int64

	for _, k := range f.keys {
		n1, err = WriteMetadataKeyTo(w, c, k)
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func WriteMetadataKeyTo(
	w io.Writer,
	c FormatterContext,
	key *catgut.String,
) (n int64, err error) {
	m := c.GetMetadata()

	var n1 int

	switch key {
	case keyAkte:
		n1, err = writeShaKeyIfNotNull(
			w,
			keyAkte,
			&m.Blob,
		)

		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

	case keyBezeichnung:
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

	case keyEtikett:
		es := m.GetTags()

		if es == nil {
			break
		}

		var sortedValues []ids.Tag

		func() {
			defer func() {
				_ = recover()
			}()

			sortedValues = quiter.SortedValues(es)
		}()

		for _, e := range sortedValues {
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

	case keyKennung:
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

	case keyShasMutterMetadataKennungMutter:
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

	case keyShasMutterMetadataKennungMutter:
		n1, err = writeShaKeyIfNotNull(
			w,
			keyShasMutterMetadataKennungMutter,
			&m.Mother,
		)

		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_strings.Tai:
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

	case keyTyp:
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

	default:
		panic(fmt.Sprintf("unsupported key: %s", key))
	}

	return
}

func writeShaKeyIfNotNull(
	w io.Writer,
	key *catgut.String,
	sh *sha.Sha,
) (n int, err error) {
	if sh.IsNull() {
		return
	}

	n, err = ohio.WriteKeySpaceValueNewlineString(
		w,
		key.String(),
		sh.String(),
	)
	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func GetShaForContext(
	f FormatGeneric,
	c FormatterContext,
) (sh interfaces.BlobId, err error) {
	m := c.GetMetadata()

	switch f.key {
	case "Akte", "AkteTyp":
		if m.Blob.IsNull() {
			return
		}

	case "AkteBez":
		if m.Blob.IsNull() && m.Description.IsEmpty() {
			return
		}
	}

	if m.GetTai().IsEmpty() {
		err = ErrEmptyTai
		return
	}

	return getShaForContext(f, c)
}

func GetShaForMetadata(
	f FormatGeneric,
	m *Metadata,
) (sh interfaces.BlobId, err error) {
	return GetShaForContext(f, nopFormatterContext{m})
}

func WriteMetadata(
	w io.Writer,
	f FormatGeneric,
	c FormatterContext,
) (blobId interfaces.BlobId, err error) {
	writer, repool := blob_ids.MakeWriterWithRepool(sha.Env{}, w)
	defer repool()

	_, err = f.WriteMetadataTo(writer, c)
	if err != nil {
		err = errors.Wrap(err)
		return
	}

	blobId = writer.GetBlobId()

	return
}

func getShaForContext(
	f FormatGeneric,
	c FormatterContext,
) (sh interfaces.BlobId, err error) {
	if sh, err = WriteMetadata(nil, f, c); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func GetShaForContextDebug(
	f FormatGeneric,
	c FormatterContext,
) (blobId interfaces.BlobId, err error) {
	var sb strings.Builder
	writer, repool := blob_ids.MakeWriterWithRepool(sha.Env{}, &sb)
	defer repool()

	_, err = f.WriteMetadataTo(writer, c)
	if err != nil {
		err = errors.Wrap(err)
		return
	}

	blobId = writer.GetBlobId()

	return
}

func GetShasForMetadata(
	m *Metadata,
) (blobIds map[string]interfaces.BlobId, err error) {
	blobIds = make(map[string]interfaces.BlobId, len(FormatKeysV5))

	for _, k := range FormatKeysV5 {
		f := FormatForKey(k)

		var sh interfaces.BlobId

		if sh, err = GetShaForMetadata(f, m); err != nil {
			err = errors.Wrap(err)
			return
		}

		if sh == nil {
			continue
		}

		blobIds[k] = sh
	}

	return
}
