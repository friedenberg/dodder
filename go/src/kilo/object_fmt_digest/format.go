package object_fmt_digest

import (
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/markl_io"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/ohio"
	"code.linenisgreat.com/dodder/go/src/echo/catgut"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/key_strings"
	"code.linenisgreat.com/dodder/go/src/foxtrot/key_strings_german"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
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

	for _, key := range format.keys {
		n1, err = writeMetadataKeyTo(writer, context, key)
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}
	}

	return n, err
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
	metadata := context.GetMetadataMutable()

	var n1 int

	switch key {
	case key_strings_german.Akte, key_strings.Blob:
		n1, err = writeMarklIdKeyIfNotNull(
			writer,
			key,
			metadata.GetBlobDigestMutable(),
		)

		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

	case key_strings_german.Bezeichnung, key_strings.Description:
		lines := strings.Split(metadata.GetDescription().String(), "\n")

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
				return n, err
			}
		}

	case key_strings_german.Etikett, key_strings.Tag:
		tags := metadata.GetTags()

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
				return n, err
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
			return n, err
		}

		n1, err = ohio.WriteKeySpaceValueNewlineString(
			writer,
			key.String(),
			context.GetObjectId().String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return n, err
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
			return n, err
		}

		n1, err = ohio.WriteKeySpaceValueNewlineString(
			writer,
			key_strings_german.Kennung.String(),
			context.GetObjectId().String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

	case key_strings_german.ShasMutterMetadateiKennungMutter:
		n1, err = writeMarklIdKeyIfNotNull(
			writer,
			key_strings_german.ShasMutterMetadateiKennungMutter,
			metadata.GetMotherObjectSig(),
		)

		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

	case key_strings.ZZRepoPub:
		n1, err = writeMarklIdKeyIfNotNull(
			writer,
			key,
			metadata.GetRepoPubKey(),
		)

		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

	case key_strings.ZZSigMother:
		n1, err = writeMarklIdKeyIfNotNull(
			writer,
			key,
			metadata.GetMotherObjectSig(),
		)

		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

	case key_strings_german.ShasMutterMetadateiKennungMutter:
		n1, err = writeMarklIdKeyIfNotNull(
			writer,
			key,
			metadata.GetMotherObjectSig(),
		)

		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

	case key_strings.Tai:
		n1, err = ohio.WriteKeySpaceValueNewlineString(
			writer,
			key_strings.Tai.String(),
			metadata.GetTai().String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

	case key_strings_german.Typ, key_strings.Type:
		if !metadata.GetType().IsEmpty() {
			n1, err = ohio.WriteKeySpaceValueNewlineString(
				writer,
				key.String(),
				metadata.GetType().String(),
			)
			n += int64(n1)

			if err != nil {
				err = errors.Wrap(err)
				return n, err
			}
		}

	case key_strings.TypeLock:
		typeTuple := metadata.GetTypeTuple()

		if typeTuple.IsEmpty() {
			err = errors.Errorf("empty type tuple")
			return n, err
		}

		if typeTuple.Key.IsEmpty() {
			err = errors.Errorf("empty type")
			return n, err
		}

		if typeTuple.Value.IsEmpty() {
			err = errors.Errorf("empty type lock")
			return n, err
		}

		n1, err = ohio.WriteKeySpaceValueNewlineString(
			writer,
			key.String(),
			typeTuple.String(),
		)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

	default:
		err = errors.Errorf("unsupported key: %s", key)
		return n, err
	}

	return n, err
}

func writeMarklIdKey(
	writer io.Writer,
	key *catgut.String,
	id interfaces.MarklId,
) (n int, err error) {
	if err = markl.AssertIdIsNotNull(id); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n, err = ohio.WriteKeySpaceValueNewlineString(
		writer,
		key.String(),
		markl.FormatBytesAsHex(id),
	)
	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

func writeMarklIdKeyIfNotNull(
	writer io.Writer,
	key *catgut.String,
	id interfaces.MarklId,
) (n int, err error) {
	if id.IsNull() {
		return n, err
	}

	return writeMarklIdKey(writer, key, id)
}

func GetDigestForContext(
	format Format,
	context FormatterContext,
) (digest interfaces.MarklId, err error) {
	metadata := context.GetMetadataMutable()

	if metadata.GetTai().IsEmpty() {
		err = ErrEmptyTai
		return digest, err
	}

	if digest, err = WriteMetadata(nil, format, context); err != nil {
		err = errors.Wrap(err)
		return digest, err
	}

	return digest, err
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
		return blobDigest, err
	}

	blobDigest = marklWriter.GetMarklId()

	return blobDigest, err
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
		return digest, err
	}

	digest = writer.GetMarklId()

	value := sb.String()

	ui.Debug().Printf("%q -> %s", value, markl.FormatBytesAsHex(digest))

	return digest, err
}
