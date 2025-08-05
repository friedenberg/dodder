package options_print

type V1 struct {
	Abbreviations abbreviationsV1 `toml:"abbreviations"`
	boxV1

	PrintMatchedDormant bool `toml:"print-matched-dormant"`
	PrintShas           bool `toml:"print-shas"`
	PrintFlush          bool `toml:"print-flush"`
	PrintUnchanged      bool `toml:"print-unchanged"`
	PrintColors         bool `toml:"print-colors"`
	PrintInventoryLists bool `toml:"print-inventory_lists"`
}

type abbreviationsV1 struct {
	ZettelIds bool `toml:"zettel-ids"`
	Shas      bool `toml:"shas"`
}

func (abbreviations abbreviationsV1) GetAbbreviations() Abbreviations {
	return Abbreviations(abbreviations)
}

type boxV1 struct {
	PrintIncludeDescription bool `toml:"print-include-description"`
	PrintTime               bool `toml:"print-time"`
	PrintTagsAlways         bool `toml:"print-etiketten-always"`
	PrintEmptyShas          bool `toml:"print-empty-shas"`
	PrintIncludeTypes       bool `toml:"print-include-typen"`
}

func (box boxV1) GetBox() Box {
	return Box{
		PrintIncludeDescription: box.PrintIncludeDescription,
		PrintTime:               box.PrintTime,
		PrintTagsAlways:         box.PrintTagsAlways,
		PrintEmptyShas:          box.PrintEmptyShas,
		PrintIncludeTypes:       box.PrintIncludeTypes,
	}
}

func (blob V1) GetPrintOptions() Options {
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
