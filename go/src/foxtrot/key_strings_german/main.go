package key_strings_german

import "code.linenisgreat.com/dodder/go/src/echo/catgut"

var (
	Akte        = catgut.MakeFromString("Akte")
	Bezeichnung = catgut.MakeFromString("Bezeichnung")
	Etikett     = catgut.MakeFromString("Etikett")
	Gattung     = catgut.MakeFromString("Gattung")
	Kennung     = catgut.MakeFromString("Kennung")
	Komment     = catgut.MakeFromString("Komment")
	Typ         = catgut.MakeFromString("Typ")

	ShasMutterMetadateiKennungMutter = catgut.MakeFromString(
		"ShasMutterMetadateiKennungMutter",
	)

	VerzeichnisseArchiviert = catgut.MakeFromString(
		"Verzeichnisse-Archiviert",
	)
	VerzeichnisseEtikettImplicit = catgut.MakeFromString(
		"Verzeichnisse-Etikett-Implicit",
	)
	VerzeichnisseEtikettExpanded = catgut.MakeFromString(
		"Verzeichnisse-Etikett-Expanded",
	)
)
