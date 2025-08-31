package object_inventory_format

import (
	"fmt"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/markl_io"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/key_bytes"
	"code.linenisgreat.com/dodder/go/src/delta/key_strings"
	"code.linenisgreat.com/dodder/go/src/delta/key_strings_german"
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
		n1, err = writeMetadataKeyTo(writer, context, k)
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func writeMetadataKeyTo(
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
			m.GetMotherObjectSig(),
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
			m.GetMotherObjectSig(),
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
			m.GetMotherObjectSig(),
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
	merkleId interfaces.MarklId,
) (n int, err error) {
	if err = markl.MakeErrIsNull(merkleId, key.String()); err != nil {
		err = errors.Wrap(err)
		return
	}

	n, err = ohio.WriteKeySpaceValueNewlineString(
		w,
		key.String(),
		markl.Format(merkleId),
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
	merkleId interfaces.MarklId,
) (n int, err error) {
	if merkleId.IsNull() {
		return
	}

	return writeMerkleIdKey(w, key, merkleId)
}

func GetDigestForContext(
	format Format,
	context FormatterContext,
) (digest interfaces.MarklId, err error) {
	m := context.GetMetadata()

	switch format.key {
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

	return getDigestForContext(format, context)
}

func GetDigestForMetadata(
	format Format,
	metadata *object_metadata.Metadata,
) (digest interfaces.MarklId, err error) {
	return GetDigestForContext(format, nopFormatterContext{metadata})
}

func WriteMetadata(
	writer io.Writer,
	format Format,
	context FormatterContext,
) (blobDigest interfaces.MarklId, err error) {
	marklWriter, repool := markl_io.MakeWriterWithRepool(
		markl.HashTypeSha256.Get(),
		writer,
	)
	defer repool()

	_, err = format.WriteMetadataTo(marklWriter, context)
	if err != nil {
		err = errors.Wrap(err)
		return
	}

	blobDigest = marklWriter.GetMarklId()

	return
}

func getDigestForContext(
	format Format,
	context FormatterContext,
) (digest interfaces.MarklId, err error) {
	if digest, err = WriteMetadata(nil, format, context); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func GetDigestForContextDebug(
	format Format,
	context FormatterContext,
) (digest interfaces.MarklId, err error) {
	var sb strings.Builder
	writer, repool := markl_io.MakeWriterWithRepool(
		markl.HashTypeSha256.Get(),
		&sb,
	)
	defer repool()

	_, err = format.WriteMetadataTo(writer, context)
	if err != nil {
		err = errors.Wrap(err)
		return
	}

	digest = writer.GetMarklId()

	value := sb.String()

	ui.Debug().Printf("%q -> %s", value, markl.Format(digest))

	return
}

func GetDigestsForMetadata(
	metadata *object_metadata.Metadata,
) (digests map[string]interfaces.MarklId, err error) {
	digests = make(map[string]interfaces.MarklId, len(FormatKeysV5))

	for _, k := range FormatKeysV5 {
		f := FormatForKey(k)

		var digest interfaces.MarklId

		if digest, err = GetDigestForMetadata(f, metadata); err != nil {
			err = errors.Wrap(err)
			return
		}

		if digest == nil {
			continue
		}

		digests[k] = digest
	}

	return
}
