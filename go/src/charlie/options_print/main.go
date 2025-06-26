package options_print

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/bravo/values"
)

func Default() V0 {
	return V0{
		Abbreviations: Abbreviations{
			ZettelIds: true,
			Shas:      true,
		},
		BoxV0: BoxV0{
			PrintIncludeTypes:       true,
			PrintIncludeDescription: true,
			PrintTime:               true,
			PrintTagsAlways:         true,
			PrintEmptyShas:          false,
		},
		PrintMatchedDormant: false,
		PrintShas:           true,
		PrintFlush:          true,
		PrintUnchanged:      true,
		PrintColors:         true,
		PrintInventoryLists: true,
	}
}

func boolVarWithMask(
	f *flag.FlagSet,
	name string,
	v *bool,
	m *bool,
	desc string,
) {
	f.Func(name,
		desc,
		func(value string) (err error) {
			var bv values.Bool

			*m = true

			if err = bv.Set(value); err != nil {
				return
			}

			*v = bv.Bool()

			return
		},
	)
}

func (c *V0) AddToFlags(f *flag.FlagSet, m *V0) {
	boolVarWithMask(
		f,
		"print-typen",
		&c.PrintIncludeTypes,
		&m.PrintIncludeTypes,
		"",
	)

	// TODO-P4 combine below three options
	boolVarWithMask(
		f,
		"abbreviate-shas",
		&c.Abbreviations.Shas,
		&m.Abbreviations.Shas,
		"",
	)

	boolVarWithMask(
		f,
		"abbreviate-zettel-ids",
		&c.Abbreviations.ZettelIds,
		&m.Abbreviations.ZettelIds,
		"",
	)

	boolVarWithMask(
		f,
		"print-description",
		&c.PrintIncludeDescription,
		&m.PrintIncludeDescription,
		"",
	)

	boolVarWithMask(
		f,
		"print-time",
		&c.PrintTime,
		&m.PrintTime,
		"",
	)

	boolVarWithMask(
		f,
		"print-etiketten",
		&c.PrintTagsAlways,
		&m.PrintTagsAlways,
		"",
	)

	boolVarWithMask(
		f,
		"print-empty-shas",
		&c.PrintEmptyShas,
		&m.PrintEmptyShas,
		"",
	)

	boolVarWithMask(
		f,
		"print-matched-archiviert",
		&c.PrintMatchedDormant,
		&m.PrintMatchedDormant,
		"",
	)

	boolVarWithMask(
		f,
		"print-shas",
		&c.PrintShas,
		&m.PrintShas,
		"",
	)

	boolVarWithMask(
		f,
		"print-flush",
		&c.PrintFlush,
		&m.PrintFlush,
		"",
	)

	boolVarWithMask(
		f,
		"print-unchanged",
		&c.PrintUnchanged,
		&m.PrintUnchanged,
		"",
	)

	boolVarWithMask(
		f,
		"print-colors",
		&c.PrintColors,
		&m.PrintColors,
		"",
	)

	boolVarWithMask(
		f,
		"print-inventory_list",
		&c.PrintInventoryLists,
		&m.PrintInventoryLists,
		"",
	)

	boolVarWithMask(
		f,
		"boxed-description",
		&c.DescriptionInBox,
		&m.DescriptionInBox,
		"",
	)
}
