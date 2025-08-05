package options_print

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/bravo/values"
)

type (
	Abbreviations struct {
		ZettelIds bool
		Shas      bool
	}

	Box struct {
		PrintIncludeDescription bool
		PrintTime               bool
		PrintTagsAlways         bool
		PrintEmptyShas          bool
		PrintIncludeTypes       bool
		PrintTai                bool
		DescriptionInBox        bool
		ExcludeFields           bool
	}

	Options struct {
		Abbreviations Abbreviations
		Box

		PrintMatchedDormant bool
		PrintShas           bool
		PrintFlush          bool
		PrintUnchanged      bool
		PrintColors         bool
		PrintInventoryLists bool
		Newlines            bool
	}

	Getter interface {
		GetPrintOptions() Options
	}
)

var (
	_ Getter = V0{}
	_ Getter = V1{}
)

func Default() V1 {
	return V1{
		Abbreviations: abbreviationsV1{
			ZettelIds: true,
			Shas:      true,
		},
		boxV1: boxV1{
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

func (dst Options) WithPrintShas(v bool) Options {
	dst.PrintShas = v
	return dst
}

func (dst Options) WithDescriptionInBox(v bool) Options {
	dst.DescriptionInBox = v
	return dst
}

func (dst Options) WithPrintTai(v bool) Options {
	dst.PrintTai = v
	return dst
}

func (dst Options) WithExcludeFields(v bool) Options {
	dst.ExcludeFields = v
	return dst
}

func (dst Options) WithPrintTime(v bool) Options {
	dst.PrintTime = v
	return dst
}

func boolVarWithMask(
	flagSet *flag.FlagSet,
	name string,
	valuePtr *bool,
	mask *bool,
	desc string,
) {
	flagSet.Func(name,
		desc,
		func(value string) (err error) {
			var bv values.Bool

			*mask = true

			if err = bv.Set(value); err != nil {
				return
			}

			*valuePtr = bv.Bool()

			return
		},
	)
}

func (dst *Options) Merge(src Options, mask Options) {
	if mask.Abbreviations.ZettelIds {
		dst.Abbreviations.ZettelIds = src.Abbreviations.ZettelIds
	}

	if mask.Abbreviations.Shas {
		dst.Abbreviations.Shas = src.Abbreviations.Shas
	}

	if mask.PrintIncludeTypes {
		dst.PrintIncludeTypes = src.PrintIncludeTypes
	}

	if mask.PrintIncludeDescription {
		dst.PrintIncludeDescription = src.PrintIncludeDescription
	}

	if mask.PrintTime {
		dst.PrintTime = src.PrintTime
	}

	if mask.PrintTagsAlways {
		dst.PrintTagsAlways = src.PrintTagsAlways
	}

	if mask.PrintEmptyShas {
		dst.PrintEmptyShas = src.PrintEmptyShas
	}

	if mask.PrintMatchedDormant {
		dst.PrintMatchedDormant = src.PrintMatchedDormant
	}

	if mask.PrintShas {
		dst.PrintShas = src.PrintShas
	}

	if mask.PrintFlush {
		dst.PrintFlush = src.PrintFlush
	}

	if mask.PrintUnchanged {
		dst.PrintUnchanged = src.PrintUnchanged
	}

	if mask.PrintColors {
		dst.PrintColors = src.PrintColors
	}

	if mask.PrintInventoryLists {
		dst.PrintInventoryLists = src.PrintInventoryLists
	}

	dst.Newlines = src.Newlines
}

// TODO rename flags away from german
func (dst *Options) AddToFlags(flagSet *flag.FlagSet, mask *Options) {
	boolVarWithMask(
		flagSet,
		"print-types",
		&dst.PrintIncludeTypes,
		&mask.PrintIncludeTypes,
		"",
	)

	// TODO-P4 combine below three options
	boolVarWithMask(
		flagSet,
		"abbreviate-shas",
		&dst.Abbreviations.Shas,
		&mask.Abbreviations.Shas,
		"",
	)

	boolVarWithMask(
		flagSet,
		"abbreviate-zettel-ids",
		&dst.Abbreviations.ZettelIds,
		&mask.Abbreviations.ZettelIds,
		"",
	)

	boolVarWithMask(
		flagSet,
		"print-description",
		&dst.PrintIncludeDescription,
		&mask.PrintIncludeDescription,
		"",
	)

	boolVarWithMask(
		flagSet,
		"print-time",
		&dst.PrintTime,
		&mask.PrintTime,
		"",
	)

	boolVarWithMask(
		flagSet,
		"print-tags",
		&dst.PrintTagsAlways,
		&mask.PrintTagsAlways,
		"",
	)

	boolVarWithMask(
		flagSet,
		"print-empty-shas",
		&dst.PrintEmptyShas,
		&mask.PrintEmptyShas,
		"",
	)

	boolVarWithMask(
		flagSet,
		"print-matched-dormant",
		&dst.PrintMatchedDormant,
		&mask.PrintMatchedDormant,
		"",
	)

	boolVarWithMask(
		flagSet,
		"print-shas",
		&dst.PrintShas,
		&mask.PrintShas,
		"",
	)

	boolVarWithMask(
		flagSet,
		"print-flush",
		&dst.PrintFlush,
		&mask.PrintFlush,
		"",
	)

	boolVarWithMask(
		flagSet,
		"print-unchanged",
		&dst.PrintUnchanged,
		&mask.PrintUnchanged,
		"",
	)

	boolVarWithMask(
		flagSet,
		"print-colors",
		&dst.PrintColors,
		&mask.PrintColors,
		"",
	)

	boolVarWithMask(
		flagSet,
		"print-inventory_list",
		&dst.PrintInventoryLists,
		&mask.PrintInventoryLists,
		"",
	)

	boolVarWithMask(
		flagSet,
		"boxed-description",
		&dst.DescriptionInBox,
		&mask.DescriptionInBox,
		"",
	)
}
