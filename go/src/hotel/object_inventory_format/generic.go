package object_inventory_format

import (
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
	"code.linenisgreat.com/dodder/go/src/delta/key_strings"
	"code.linenisgreat.com/dodder/go/src/delta/key_strings_german"
)

const (
	KeyFormatV5Metadata               = "Metadatei"
	KeyFormatV5MetadataWithoutTai     = "MetadateiSansTai"
	KeyFormatV5MetadataObjectIdParent = "MetadateiKennungMutter"
	KeyFormatV11ObjectDigest          = merkle.HRPObjectDigestSha256V1
)

var (
	FormatKeys = []string{
		KeyFormatV5Metadata,
		KeyFormatV5MetadataWithoutTai,
		KeyFormatV5MetadataObjectIdParent,
	}

	FormatKeysV5 = []string{
		KeyFormatV5Metadata,
		KeyFormatV5MetadataWithoutTai,
		KeyFormatV5MetadataObjectIdParent,
	}

	FormatsV5MetadataSansTai = Format{
		key: KeyFormatV5MetadataWithoutTai,
		keys: []keyType{
			keyAkte,
			keyBezeichnung,
			keyEtikett,
			keyTyp,
		},
	}

	FormatV11ObjectDigest = Format{
		key: KeyFormatV11ObjectDigest,
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
	metadataSansTai        Format
	metadata               Format
	metadataObjectIdParent Format
}

func (formats formats) MetadataSansTai() Format {
	return formats.metadataSansTai
}

func (formats formats) Metadata() Format {
	return formats.metadata
}

func (formats formats) MetadataObjectIdParent() Format {
	return formats.metadataObjectIdParent
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

	Formats.metadataSansTai.key = KeyFormatV5MetadataWithoutTai
	Formats.metadataSansTai.keys = []keyType{
		keyAkte,
		keyBezeichnung,
		keyEtikett,
		keyTyp,
	}

	Formats.metadataObjectIdParent.key = KeyFormatV5MetadataObjectIdParent
	Formats.metadataObjectIdParent.keys = []keyType{
		keyAkte,
		keyBezeichnung,
		keyEtikett,
		keyKennung,
		keyTyp,
		key_strings.Tai,
		keyShasMutterMetadataKennungMutter,
	}
}
