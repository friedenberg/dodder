package cli

import (
	"flag"
	"maps"
	"slices"
	"strings"
)

// TODO add support for comma-separated values
type CLICompleter interface {
	GetCLICompletion() map[string]string
}

type FlagValueWithCompetion interface {
	flag.Value
	CLICompleter
}

func FlagSetVarWithCompletion(
	flagSet *flag.FlagSet,
	value FlagValueWithCompetion,
	key string,
) {
	flagSet.Var(
		value,
		key,
		strings.Join(
			slices.Collect(
				maps.Keys(value.GetCLICompletion()),
			),
			", ",
		),
	)
}
