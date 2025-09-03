// TODO rename
package object_inventory_format

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/key_strings"
	"code.linenisgreat.com/dodder/go/src/delta/key_strings_german"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
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

	nopFormatterContext struct {
		object_metadata.PersistentFormatterContext
	}
)

func (nopFormatterContext) GetObjectId() *ids.ObjectId {
	return nil
}

type keyType = *catgut.String

type Format struct {
	marklFormatId string
	keys          []keyType
}

func (format Format) GetMarklTypeId() string {
	return format.marklFormatId
}

func FormatForMarklFormatIdError(
	marklFormatId string,
) (format Format, err error) {
	var found bool
	if format, found = formatsMap[marklFormatId]; !found {
		err = errUnknownFormatKey(marklFormatId)
		return
	}

	return
}

var (
	formatsList = []Format{}
	formatsMap  = map[string]Format{}
)

func registerFormat(marklFormatId string, keys ...keyType) {
	format, alreadyExists := formatsMap[marklFormatId]

	if alreadyExists {
		panic(
			fmt.Sprintf(
				"format for markl format id %q already registered",
				marklFormatId,
			),
		)
	}

	format.marklFormatId = marklFormatId
	format.keys = keys

	formatsList = append(formatsList, format)
	formatsMap[marklFormatId] = format
}

func init() {
	registerFormat(
		markl.FormatIdV5MetadataDigestWithoutTai,
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
		markl.FormatIdObjectDigestSha256V1,
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
