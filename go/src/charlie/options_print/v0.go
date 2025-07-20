package options_print

type V0 struct {
	Abbreviations abbreviationsV0 `toml:"abbreviations"`
	boxV0

	PrintMatchedDormant bool `toml:"print-matched-archiviert"`
	PrintShas           bool `toml:"print-shas"`
	PrintFlush          bool `toml:"print-flush"`
	PrintUnchanged      bool `toml:"print-unchanged"`
	PrintColors         bool `toml:"print-colors"`
	PrintInventoryLists bool `toml:"print-bestandsaufnahme"`
}

type abbreviationsV0 struct {
	ZettelIds bool `toml:"hinweisen"`
	Shas      bool `toml:"shas"`
}

func (abbreviations abbreviationsV0) GetAbbreviations() Abbreviations {
	return Abbreviations{
		ZettelIds: abbreviations.ZettelIds,
		Shas:      abbreviations.Shas,
	}
}

type boxV0 struct {
	PrintIncludeDescription bool `toml:"print-include-description"`
	PrintTime               bool `toml:"print-time"`
	PrintTagsAlways         bool `toml:"print-etiketten-always"`
	PrintEmptyShas          bool `toml:"print-empty-shas"`
	PrintIncludeTypes       bool `toml:"print-include-typen"`
}

func (box boxV0) GetBox() Box {
	return Box{
		PrintIncludeDescription: box.PrintIncludeDescription,
		PrintTime:               box.PrintTime,
		PrintTagsAlways:         box.PrintTagsAlways,
		PrintEmptyShas:          box.PrintEmptyShas,
		PrintIncludeTypes:       box.PrintIncludeTypes,
	}
}

func (blob V0) GetPrintOptions() Options {
	return Options{
		Abbreviations:       blob.Abbreviations.GetAbbreviations(),
		Box:                 blob.GetBox(),
		PrintMatchedDormant: blob.PrintMatchedDormant,
		PrintShas:           blob.PrintShas,
		PrintFlush:          blob.PrintFlush,
		PrintUnchanged:      blob.PrintUnchanged,
		PrintColors:         blob.PrintColors,
		PrintInventoryLists: blob.PrintInventoryLists,
	}
}
