package options_print

type V0 struct {
	Abbreviations Abbreviations `toml:"abbreviations"`
	BoxV0
	PrintMatchedDormant bool `toml:"print-matched-archiviert"`
	PrintShas           bool `toml:"print-shas"`
	PrintFlush          bool `toml:"print-flush"`
	PrintUnchanged      bool `toml:"print-unchanged"`
	PrintColors         bool `toml:"print-colors"`
	PrintInventoryLists bool `toml:"print-bestandsaufnahme"`
	ZittishNewlines     bool `toml:"-"`
}

func (a *V0) Merge(b V0, mask V0) {
	if mask.Abbreviations.ZettelIds {
		a.Abbreviations.ZettelIds = b.Abbreviations.ZettelIds
	}

	if mask.Abbreviations.Shas {
		a.Abbreviations.Shas = b.Abbreviations.Shas
	}

	if mask.PrintIncludeTypes {
		a.PrintIncludeTypes = b.PrintIncludeTypes
	}

	if mask.PrintIncludeDescription {
		a.PrintIncludeDescription = b.PrintIncludeDescription
	}

	if mask.PrintTime {
		a.PrintTime = b.PrintTime
	}

	if mask.PrintTagsAlways {
		a.PrintTagsAlways = b.PrintTagsAlways
	}

	if mask.PrintEmptyShas {
		a.PrintEmptyShas = b.PrintEmptyShas
	}

	if mask.PrintMatchedDormant {
		a.PrintMatchedDormant = b.PrintMatchedDormant
	}

	if mask.PrintShas {
		a.PrintShas = b.PrintShas
	}

	if mask.PrintFlush {
		a.PrintFlush = b.PrintFlush
	}

	if mask.PrintUnchanged {
		a.PrintUnchanged = b.PrintUnchanged
	}

	if mask.PrintColors {
		a.PrintColors = b.PrintColors
	}

	if mask.PrintInventoryLists {
		a.PrintInventoryLists = b.PrintInventoryLists
	}

	if mask.DescriptionInBox {
		a.DescriptionInBox = b.DescriptionInBox
	}

	a.ZittishNewlines = b.ZittishNewlines
}

func (c V0) WithPrintShas(v bool) V0 {
	c.PrintShas = v
	return c
}

func (c V0) WithDescriptionInBox(v bool) V0 {
	c.DescriptionInBox = v
	return c
}

func (c V0) WithPrintTai(v bool) V0 {
	c.PrintTai = v
	return c
}

func (c V0) WithExcludeFields(v bool) V0 {
	c.ExcludeFields = v
	return c
}

func (c V0) WithPrintTime(v bool) V0 {
	c.PrintTime = v
	return c
}
