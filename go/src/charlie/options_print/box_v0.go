package options_print

type BoxV0 struct {
	PrintIncludeDescription bool `toml:"print-include-description"`
	PrintTime               bool `toml:"print-time"`
	PrintTagsAlways         bool `toml:"print-etiketten-always"`
	PrintEmptyShas          bool `toml:"print-empty-shas"`
	PrintIncludeTypes       bool `toml:"print-include-typen"`

	PrintTai         bool `toml:"-"`
	DescriptionInBox bool `toml:"-"`
	ExcludeFields    bool `toml:"-"`
}
