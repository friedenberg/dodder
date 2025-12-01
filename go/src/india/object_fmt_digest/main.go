package object_fmt_digest

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/catgut"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/key_strings"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/hotel/objects"
)

type (
	FormatterContext interface {
		objects.PersistentFormatterContext
		GetObjectId() *ids.ObjectId
	}
)

type keyType = *catgut.String

func GetFormatForPurpose(
	purpose string,
) (format format) {
	var found bool

	if format, found = formatsMap[purpose]; !found {
		panic(errUnknownFormatKey(purpose))
	}

	return format
}

func FormatForPurposeOrError(
	purpose string,
) (format format, err error) {
	var found bool
	if format, found = formatsMap[purpose]; !found {
		err = errUnknownFormatKey(purpose)
		return format, err
	}

	return format, err
}

var (
	formatsList = []format{}
	formatsMap  = map[string]format{}
)

func registerFormat(purpose string, keys ...keyType) {
	format, alreadyExists := formatsMap[purpose]

	if alreadyExists {
		panic(
			fmt.Sprintf(
				"format for purpose %q already registered",
				purpose,
			),
		)
	}

	format.purpose = purpose
	format.keys = keys

	formatsList = append(formatsList, format)
	formatsMap[purpose] = format
}

// TODO register these formats with a hash of the code used to generate them to
// make them immutable. Maybe implement as plugins.
func init() {
	// TODO replace with modern keys
	registerFormat(
		markl.PurposeV5MetadataDigestWithoutTai,
		key_strings.Blob,
		key_strings.Description,
		key_strings.Tag,
		key_strings.Type,
	)

	registerFormat(
		markl.PurposeObjectDigestV1,
		key_strings.Blob,
		key_strings.Description,
		key_strings.ObjectId,
		key_strings.Tag,
		key_strings.Tai,
		key_strings.Type,
		key_strings.ZZRepoPub,
		key_strings.ZZSigMother,
	)

	registerFormat(
		markl.PurposeObjectDigestV2,
		key_strings.Blob,
		key_strings.Description,
		key_strings.ObjectId,
		key_strings.Tag,
		key_strings.Tai,
		key_strings.TypeLock,
		key_strings.ZZRepoPub,
		key_strings.ZZSigMother,
	)
}

func WriteDigest(
	formatId string,
	context FormatterContext,
	output interfaces.MarklIdMutable,
) (err error) {
	format := GetFormatForPurpose(formatId)

	metadata := context.GetMetadataMutable()

	if metadata.GetTai().IsEmpty() {
		err = ErrEmptyTai
		return err
	}

	var digest interfaces.MarklId

	if digest, err = format.writeMetadata(nil, context); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer markl.PutId(digest)

	output.ResetWithMarklId(digest)

	if err = output.SetPurposeId(format.purpose); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = markl.AssertIdIsNotNull(
		output,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
