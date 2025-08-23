package options_print

import (
	"code.linenisgreat.com/dodder/go/src/bravo/equals"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
)

type (
	OverlayAbbreviations struct {
		ZettelIds *bool
		Shas      *bool
	}

	OverlayBox struct {
		PrintIncludeDescription *bool
		PrintTime               *bool
		PrintTagsAlways         *bool
		PrintEmptyShas          *bool
		PrintIncludeTypes       *bool
		PrintTai                *bool
		DescriptionInBox        *bool
		ExcludeFields           *bool
	}

	Overlay struct {
		Abbreviations *OverlayAbbreviations
		OverlayBox

		PrintMatchedDormant *bool
		PrintShas           *bool
		PrintFlush          *bool
		PrintUnchanged      *bool
		PrintColors         *bool
		PrintInventoryLists *bool
		Newlines            *bool
	}

	Options struct {
		AbbreviateZettelIds        bool
		AbbreviateShas             bool
		BoxPrintIncludeDescription bool
		BoxPrintTime               bool
		BoxPrintTagsAlways         bool
		BoxPrintEmptyShas          bool
		BoxPrintIncludeTypes       bool
		BoxPrintTai                bool
		BoxDescriptionInBox        bool
		BoxExcludeFields           bool
		PrintMatchedDormant        bool
		PrintShas                  bool
		PrintFlush                 bool
		PrintUnchanged             bool
		PrintColors                bool
		PrintInventoryLists        bool
		Newlines                   bool
	}

	OverlayGetter interface {
		GetPrintOptionsOverlay() Overlay
	}

	OptionGetter interface {
		GetPrintOptions() Options
	}
)

var (
	_ OverlayGetter = Overlay{}
	_ OverlayGetter = V0{}
	_ OverlayGetter = V1{}
)

func Default() Options {
	return Options{
		AbbreviateZettelIds:        true,
		AbbreviateShas:             true,
		BoxPrintIncludeTypes:       true,
		BoxPrintIncludeDescription: true,
		BoxPrintTime:               true,
		BoxPrintTagsAlways:         true,
		BoxPrintEmptyShas:          false,
		PrintMatchedDormant:        false,
		PrintShas:                  true,
		PrintFlush:                 true,
		PrintUnchanged:             true,
		PrintColors:                true,
		PrintInventoryLists:        true,
	}
}

func DefaultOverlay() V1 {
	config := Default()

	return V1{
		Abbreviations: &abbreviationsV1{
			ZettelIds: &config.AbbreviateZettelIds,
			Shas:      &config.AbbreviateShas,
		},
		boxV1: boxV1{
			PrintIncludeTypes:       &config.BoxPrintIncludeTypes,
			PrintIncludeDescription: &config.BoxPrintIncludeDescription,
			PrintTime:               &config.BoxPrintTime,
			PrintTagsAlways:         &config.BoxPrintTagsAlways,
			PrintEmptyShas:          &config.BoxPrintEmptyShas,
		},
		PrintMatchedDormant: &config.PrintMatchedDormant,
		PrintShas:           &config.PrintShas,
		PrintFlush:          &config.PrintFlush,
		PrintUnchanged:      &config.PrintUnchanged,
		PrintColors:         &config.PrintColors,
		PrintInventoryLists: &config.PrintInventoryLists,
	}
}

func (options Options) WithPrintShas(v bool) Options {
	options.PrintShas = v
	return options
}

func (options Options) WithDescriptionInBox(v bool) Options {
	options.BoxDescriptionInBox = v
	return options
}

func (options Options) WithPrintTai(v bool) Options {
	options.BoxPrintTai = v
	return options
}

func (options Options) WithExcludeFields(v bool) Options {
	options.BoxExcludeFields = v
	return options
}

func (options Options) WithPrintTime(v bool) Options {
	options.BoxPrintTime = v
	return options
}

func (options Options) UsePrintTime() bool {
	return options.BoxPrintTime
}

func (options Options) UsePrintTags() bool {
	return options.BoxPrintTagsAlways
}

func MakeDefaultConfig(overlays ...OverlayGetter) Options {
	return MakeConfig(Default(), overlays...)
}

func MakeConfig(base Options, overlays ...OverlayGetter) Options {
	for _, overlayGetter := range overlays {
		overlay := overlayGetter.GetPrintOptionsOverlay()
		if abbreviations := overlay.Abbreviations; abbreviations != nil {
			equals.SetIfValueNotNil(
				&base.AbbreviateZettelIds,
				abbreviations.ZettelIds,
			)
			equals.SetIfValueNotNil(&base.AbbreviateShas, abbreviations.Shas)
		}

		box := overlay.OverlayBox
		equals.SetIfValueNotNil(&base.BoxDescriptionInBox, box.DescriptionInBox)
		equals.SetIfValueNotNil(&base.BoxPrintTime, box.PrintTime)
		equals.SetIfValueNotNil(&base.BoxPrintTagsAlways, box.PrintTagsAlways)
		equals.SetIfValueNotNil(&base.BoxPrintEmptyShas, box.PrintEmptyShas)
		equals.SetIfValueNotNil(
			&base.BoxPrintIncludeTypes,
			box.PrintIncludeTypes,
		)
		equals.SetIfValueNotNil(&base.BoxPrintTai, box.PrintTai)
		equals.SetIfValueNotNil(
			&base.BoxPrintIncludeDescription,
			box.PrintIncludeDescription,
		)
		equals.SetIfValueNotNil(&base.BoxExcludeFields, box.ExcludeFields)

		equals.SetIfValueNotNil(
			&base.PrintMatchedDormant,
			overlay.PrintMatchedDormant,
		)
		equals.SetIfValueNotNil(&base.PrintShas, overlay.PrintShas)
		equals.SetIfValueNotNil(&base.PrintFlush, overlay.PrintFlush)
		equals.SetIfValueNotNil(&base.PrintUnchanged, overlay.PrintUnchanged)
		equals.SetIfValueNotNil(&base.PrintColors, overlay.PrintColors)
		equals.SetIfValueNotNil(
			&base.PrintInventoryLists,
			overlay.PrintInventoryLists,
		)
		equals.SetIfValueNotNil(&base.Newlines, overlay.Newlines)
	}
	return base
}

func (overlay Overlay) GetPrintOptionsOverlay() Overlay {
	return overlay
}

func makeFlagSetFuncBoolVar(valuePtr **bool) func(value string) (err error) {
	return func(value string) (err error) {
		var boolValue values.Bool

		if err = boolValue.Set(value); err != nil {
			return
		}

		booll := boolValue.Bool()

		*valuePtr = &booll

		return
	}
}

func (overlay *Overlay) AddToFlags(flagSet *flags.FlagSet) {
	flagSet.Func(
		"print-types",
		"",
		makeFlagSetFuncBoolVar(&overlay.PrintIncludeTypes),
	)

	// TODO-P4 combine below three options
	flagSet.Func(
		"abbreviate-shas",
		"",
		func(value string) (err error) {
			if overlay.Abbreviations == nil {
				overlay.Abbreviations = &OverlayAbbreviations{}
			}

			return makeFlagSetFuncBoolVar(&overlay.Abbreviations.Shas)(value)
		},
	)

	flagSet.Func(
		"abbreviate-zettel-ids",
		"",
		func(value string) (err error) {
			if overlay.Abbreviations == nil {
				overlay.Abbreviations = &OverlayAbbreviations{}
			}

			return makeFlagSetFuncBoolVar(
				&overlay.Abbreviations.ZettelIds,
			)(
				value,
			)
		},
	)

	flagSet.Func(
		"print-description",
		"",
		makeFlagSetFuncBoolVar(&overlay.PrintIncludeDescription),
	)

	flagSet.Func(
		"print-time",
		"",
		makeFlagSetFuncBoolVar(&overlay.PrintTime),
	)

	flagSet.Func(
		"print-tags",
		"",
		makeFlagSetFuncBoolVar(&overlay.PrintTagsAlways),
	)

	flagSet.Func(
		"print-empty-shas",
		"",
		makeFlagSetFuncBoolVar(&overlay.PrintEmptyShas),
	)

	flagSet.Func(
		"print-matched-dormant",
		"",
		makeFlagSetFuncBoolVar(&overlay.PrintMatchedDormant),
	)

	flagSet.Func(
		"print-shas",
		"",
		makeFlagSetFuncBoolVar(&overlay.PrintShas),
	)

	flagSet.Func(
		"print-flush",
		"",
		makeFlagSetFuncBoolVar(&overlay.PrintFlush),
	)

	flagSet.Func(
		"print-unchanged",
		"",
		makeFlagSetFuncBoolVar(&overlay.PrintUnchanged),
	)

	flagSet.Func(
		"print-colors",
		"",
		makeFlagSetFuncBoolVar(&overlay.PrintColors),
	)

	flagSet.Func(
		"print-inventory_list",
		"",
		makeFlagSetFuncBoolVar(&overlay.PrintInventoryLists),
	)

	flagSet.Func(
		"boxed-description",
		"",
		makeFlagSetFuncBoolVar(&overlay.DescriptionInBox),
	)

	flagSet.Func(
		"zittish-newlines",
		"add extra newlines to zittish to improve readability",
		makeFlagSetFuncBoolVar(&overlay.Newlines),
	)
}
