package object_fmt_digest

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/catgut"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/key_strings"
	"code.linenisgreat.com/dodder/go/src/foxtrot/key_strings_german"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/juliett/object_metadata"
)

type (
	FormatterContext interface {
		object_metadata.PersistentFormatterContext
		GetObjectId() *ids.ObjectId
	}

	ParserContext interface {
		object_metadata.PersistentParserContext
		SetObjectIdLike(interfaces.ObjectId) error
	}
)

type keyType = *catgut.String

type Format struct {
	purpose string
	keys    []keyType
}

func (format Format) GetPurpose() string {
	return format.purpose
}

func FormatForPurposeOrError(
	purpose string,
) (format Format, err error) {
	var found bool
	if format, found = formatsMap[purpose]; !found {
		err = errUnknownFormatKey(purpose)
		return format, err
	}

	return format, err
}

var (
	formatsList = []Format{}
	formatsMap  = map[string]Format{}
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

func init() {
	registerFormat(
		markl.PurposeV5MetadataDigestWithoutTai,
		key_strings_german.Akte,
		key_strings_german.Bezeichnung,
		key_strings_german.Etikett,
		key_strings_german.Typ,
	)

	// registerFormat(
	// 	markl.FormatIdObjectDigestObjectId,
	// 	key_strings.ObjectId,
	// )

	// registerFormat(
	// 	markl.FormatIdObjectDigestObjectIdTai,
	// 	key_strings.ObjectId,
	// 	key_strings.Tai,
	// )

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
}
