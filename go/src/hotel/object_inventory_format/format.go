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
	"code.linenisgreat.com/dodder/go/src/delta/key_strings"
	"code.linenisgreat.com/dodder/go/src/delta/key_strings_german"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

func FormatForKey(k string) Format {
	format, err := FormatForPurposeOrError(k)
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
			key_strings_german.Gattung.String(),
			context.GetObjectId().GetGenre().GetGenreString(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		n1, err = ohio.WriteKeySpaceValueNewlineString(
			writer,
			key_strings_german.Kennung.String(),
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
			key_strings_german.ShasMutterMetadateiKennungMutter,
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
			key,
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
	if err = markl.AssertIdIsNotNull(merkleId, key.String()); err != nil {
		err = errors.Wrap(err)
		return
	}

	n, err = ohio.WriteKeySpaceValueNewlineString(
		w,
		key.String(),
		markl.FormatBytesAsHext(merkleId),
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

	if m.GetTai().IsEmpty() {
		err = ErrEmptyTai
		return
	}

	if digest, err = WriteMetadata(nil, format, context); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func WriteMetadata(
	writer io.Writer,
	format Format,
	context FormatterContext,
) (blobDigest interfaces.MarklId, err error) {
	marklWriter, repool := markl_io.MakeWriterWithRepool(
		markl.FormatHashSha256.Get(),
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

func GetDigestForContextDebug(
	format Format,
	context FormatterContext,
) (digest interfaces.MarklId, err error) {
	var sb strings.Builder
	writer, repool := markl_io.MakeWriterWithRepool(
		markl.FormatHashSha256.Get(),
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

	ui.Debug().Printf("%q -> %s", value, markl.FormatBytesAsHext(digest))

	return
}
