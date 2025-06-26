package german_keys

import "code.linenisgreat.com/dodder/go/src/delta/catgut"

var (
	KeyAkte        = catgut.MakeFromString("Akte")
	KeyBezeichnung = catgut.MakeFromString("Bezeichnung")
	KeyEtikett     = catgut.MakeFromString("Etikett")
	KeyGattung     = catgut.MakeFromString("Gattung")
	KeyKennung     = catgut.MakeFromString("Kennung")
	KeyKomment     = catgut.MakeFromString("Komment")
	KeyTyp         = catgut.MakeFromString("Typ")

	KeyShasMutterMetadateiKennungMutter = catgut.MakeFromString("ShasMutterMetadateiKennungMutter")

	KeyVerzeichnisseArchiviert      = catgut.MakeFromString("Verzeichnisse-Archiviert")
	KeyVerzeichnisseEtikettImplicit = catgut.MakeFromString("Verzeichnisse-Etikett-Implicit")
	KeyVerzeichnisseEtikettExpanded = catgut.MakeFromString("Verzeichnisse-Etikett-Expanded")
)
