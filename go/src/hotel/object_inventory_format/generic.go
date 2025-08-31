package object_inventory_format

import (
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/key_strings"
	"code.linenisgreat.com/dodder/go/src/delta/key_strings_german"
)

const (
	KeyFormatV5Metadata               = "Metadatei"
	KeyFormatV5MetadataWithoutTai     = markl.MarklTypeIdV5MetadataDigestWithoutTai
	KeyFormatV5MetadataObjectIdParent = "MetadateiKennungMutter"
)

var (
	FormatKeysV5 = []string{
		KeyFormatV5Metadata,
		KeyFormatV5MetadataWithoutTai,
		KeyFormatV5MetadataObjectIdParent,
	}

	formatV5MetadataWithoutTai = Format{
		key: markl.MarklTypeIdV5MetadataDigestWithoutTai,
		keys: []keyType{
			keyAkte,
			keyBezeichnung,
			keyEtikett,
			keyTyp,
		},
	}

	formatV11ObjectDigest = Format{
		key: markl.MarklTypeIdObjectDigestSha256V1,
		keys: []keyType{
			key_strings.Blob,
			key_strings.Description,
			key_strings.ObjectId,
			key_strings.Tag,
			key_strings.Tai,
			key_strings.Type,
			key_strings.ZZRepoPub,
			key_strings.ZZSigMother,
		},
	}

	// TODO remove local aliases and only use german_keys
	keyAkte                            = key_strings_german.Akte
	keyBezeichnung                     = key_strings_german.Bezeichnung
	keyEtikett                         = key_strings_german.Etikett
	keyGattung                         = key_strings_german.Gattung
	keyKennung                         = key_strings_german.Kennung
	keyTyp                             = key_strings_german.Typ
	keyShasMutterMetadataKennungMutter = key_strings_german.ShasMutterMetadateiKennungMutter
	keyVerzeichnisseArchiviert         = key_strings_german.VerzeichnisseArchiviert
	keyVerzeichnisseEtikettImplicit    = key_strings_german.VerzeichnisseEtikettImplicit
	keyVerzeichnisseEtikettExpanded    = key_strings_german.VerzeichnisseEtikettExpanded
)

type formats struct {
	metadata Format
}

func (formats formats) Metadata() Format {
	return formats.metadata
}

// TODO remove
var Formats formats

func init() {
	Formats.metadata.key = KeyFormatV5Metadata
	Formats.metadata.keys = []keyType{
		keyAkte,
		keyBezeichnung,
		keyEtikett,
		keyTyp,
		key_strings.Tai,
	}
}
