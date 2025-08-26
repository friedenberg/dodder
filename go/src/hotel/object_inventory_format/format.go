package object_inventory_format

import (
	"fmt"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/key_bytes"
	"code.linenisgreat.com/dodder/go/src/delta/key_strings"
	"code.linenisgreat.com/dodder/go/src/delta/key_strings_german"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

type keyType = *catgut.String

type Format struct {
	key  string
	keys []keyType
}

func FormatForKeyError(key string) (format Format, err error) {
	switch key {
	case KeyFormatV5Metadata:
		format = Formats.metadata

	case KeyFormatV5MetadataWithoutTai:
		format = Formats.metadataSansTai

	case KeyFormatV5MetadataObjectIdParent:
		format = Formats.metadataObjectIdParent

	default:
		err = errInvalidGenericFormat(key)
		return
	}

	return
}

func FormatForKey(k string) Format {
	format, err := FormatForKeyError(k)
	errors.PanicIfError(err)
	return format
}

func (format Format) WriteMetadataTo(
	writer io.Writer,
	context FormatterContext,
) (n int64, err error) {
	var n1 int64

	for _, k := range format.keys {
		n1, err = WriteMetadataKeyTo(writer, context, k)
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func WriteMetadataKeyTo(
	writer io.Writer,
	context FormatterContext,
	key keyType,
) (n int64, err error) {
	return writeMetadataKeyStringTo(writer, context, key)
	// switch key := key.(type) {
	// case key_bytes.Binary:
	// 	return writeMetadataKeyByteTo(writer, context, key)

	// case *catgut.String:
	// 	return writeMetadataKeyStringTo(writer, context, key)

	// default:
	// 	err = errors.Errorf("unsupported key: %T", key)
	// 	return
	// }
}

func writeMetadataKeyByteTo(
	writer io.Writer,
	context FormatterContext,
	key key_bytes.Binary,
) (n int64, err error) {
	return
}

func writeMetadataKeyStringTo(
	writer io.Writer,
	context FormatterContext,
	key *catgut.String,
) (n int64, err error) {
	m := context.GetMetadata()

	var n1 int

	switch key {
	case key_strings_german.Akte, key_strings.Blob:
		n1, err = writeMerkleIdKeyIfNotNull(
			writer,
			key,
			m.GetBlobDigestMutable(),
		)

		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_strings_german.Bezeichnung, key_strings.Description:
		lines := strings.Split(m.Description.String(), "\n")

		for _, line := range lines {
			if line == "" {
				continue
			}

			n1, err = ohio.WriteKeySpaceValueNewlineString(
				writer,
				key.String(),
				line,
			)
			n += int64(n1)

			if err != nil {
				err = errors.Wrap(err)
				return
			}
		}

	case key_strings_german.Etikett, key_strings.Tag:
		tags := m.GetTags()

		if tags == nil {
			break
		}

		var sortedValues []ids.Tag

		func() {
			defer func() {
				_ = recover()
			}()

			sortedValues = quiter.SortedValues(tags)
		}()

		for _, tag := range sortedValues {
			if tag.IsVirtual() {
				continue
			}

			n1, err = ohio.WriteKeySpaceValueNewlineString(
				writer,
				key.String(),
				tag.String(),
			)
			n += int64(n1)

			if err != nil {
				err = errors.Wrap(err)
				return
			}
		}

	case key_strings.ObjectId:
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			writer,
			key_strings.Genre.String(),
			context.GetObjectId().GetGenre().GetGenreString(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		n1, err = ohio.WriteKeySpaceValueNewlineString(
			writer,
			key.String(),
			context.GetObjectId().String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_strings_german.Kennung:
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			writer,
			keyGattung.String(),
			context.GetObjectId().GetGenre().GetGenreString(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		n1, err = ohio.WriteKeySpaceValueNewlineString(
			writer,
			keyKennung.String(),
			context.GetObjectId().String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_strings_german.ShasMutterMetadateiKennungMutter:
		n1, err = writeMerkleIdKeyIfNotNull(
			writer,
			keyShasMutterMetadataKennungMutter,
			m.GetMotherObjectDigest(),
		)

		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_strings.ZZRepoPub:
		n1, err = writeMerkleIdKey(
			writer,
			key,
			m.GetRepoPubKey(),
		)

		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_strings.ZZSigMother:
		n1, err = writeMerkleIdKeyIfNotNull(
			writer,
			key,
			m.GetMotherObjectDigest(),
		)

		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_strings_german.ShasMutterMetadateiKennungMutter:
		n1, err = writeMerkleIdKeyIfNotNull(
			writer,
			keyShasMutterMetadataKennungMutter,
			m.GetMotherObjectDigest(),
		)

		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_strings.Tai:
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			writer,
			key_strings.Tai.String(),
			m.Tai.String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_strings_german.Typ, key_strings.Type:
		if !m.Type.IsEmpty() {
			n1, err = ohio.WriteKeySpaceValueNewlineString(
				writer,
				key.String(),
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

func writeMerkleIdKey(
	w io.Writer,
	key *catgut.String,
	merkleId interfaces.MerkleId,
) (n int, err error) {
	if err = merkle_ids.MakeErrIsNull(merkleId); err != nil {
		err = errors.Wrap(err)
		return
	}

	n, err = ohio.WriteKeySpaceValueNewlineString(
		w,
		key.String(),
		merkle_ids.Format(merkleId),
	)
	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func writeMerkleIdKeyIfNotNull(
	w io.Writer,
	key *catgut.String,
	merkleId interfaces.MerkleId,
) (n int, err error) {
	if merkleId.IsNull() {
		return
	}

	return writeMerkleIdKey(w, key, merkleId)
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
	f Format,
	c FormatterContext,
) (sh interfaces.BlobId, err error) {
	m := c.GetMetadata()

	switch f.key {
	case "Akte", "AkteTyp":
		if m.GetBlobDigest().IsNull() {
			return
		}

	case "AkteBez":
		if m.GetBlobDigest().IsNull() && m.Description.IsEmpty() {
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
	f Format,
	m *object_metadata.Metadata,
) (sh interfaces.BlobId, err error) {
	return GetShaForContext(f, nopFormatterContext{m})
}

func WriteMetadata(
	w io.Writer,
	f Format,
	c FormatterContext,
) (blobId interfaces.BlobId, err error) {
	writer, repool := merkle_ids.MakeWriterWithRepool(sha.Env{}, w)
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
	f Format,
	c FormatterContext,
) (sh interfaces.BlobId, err error) {
	if sh, err = WriteMetadata(nil, f, c); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func GetShaForContextDebug(
	format Format,
	context FormatterContext,
) (blobId interfaces.BlobId, err error) {
	var sb strings.Builder
	writer, repool := merkle_ids.MakeWriterWithRepool(sha.Env{}, &sb)
	defer repool()

	_, err = format.WriteMetadataTo(writer, context)
	if err != nil {
		err = errors.Wrap(err)
		return
	}

	blobId = writer.GetBlobId()

	ui.Debug().Printf("%q -> %s", &sb, merkle_ids.Format(blobId))

	return
}

func GetShasForMetadata(
	m *object_metadata.Metadata,
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
