package repo_config_cli

import (
	"flag"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/bravo/options_tools"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/delta/debug"
	"code.linenisgreat.com/dodder/go/src/echo/descriptions"
)

type Blob struct {
	BasePath string

	Debug            debug.Options
	Verbose          bool
	Quiet            bool
	Todo             bool
	dryRun           bool
	IgnoreHookErrors bool
	Hooks            string

	CheckoutCacheEnabled bool
	PredictableZettelIds bool

	PrintOptions, maskPrintOptions options_print.Options
	ToolOptions                    options_tools.Options

	descriptions.Description
}

// TODO add support for all flags
func (config Blob) GetCLIFlags() (flags []string) {
	flags = append(
		flags,
		fmt.Sprintf("-print-time=%t", config.PrintOptions.PrintTime),
	)
	flags = append(
		flags,
		fmt.Sprintf("-print-colors=%t", config.PrintOptions.PrintColors),
	)

	if config.Verbose {
		flags = append(flags, "-verbose")
	}

	return
}

func (config *Blob) SetFlagSet(flagSet *flag.FlagSet) {
	flagSet.StringVar(&config.BasePath, "dir-dodder", "", "")

	flagSet.Var(&config.Debug, "debug", "debugging options")
	flagSet.BoolVar(&config.Todo, "todo", false, "")
	flagSet.BoolVar(&config.dryRun, "dry-run", false, "")
	flagSet.BoolVar(&config.Verbose, "verbose", false, "")
	flagSet.BoolVar(&config.Quiet, "quiet", false, "")

	flagSet.BoolVar(
		&config.CheckoutCacheEnabled,
		"checkout-cache-enabled",
		false,
		"",
	)

	flagSet.BoolVar(
		&config.PredictableZettelIds,
		"predictable-zettel-ids",
		false,
		"generate new zettel ids in order",
	)

	config.PrintOptions.AddToFlags(flagSet, &config.maskPrintOptions)
	config.ToolOptions.AddToFlags(flagSet)

	flagSet.BoolVar(
		&config.PrintOptions.Newlines,
		"zittish-newlines",
		false,
		"add extra newlines to zittish to improve readability",
	)

	flagSet.BoolVar(
		&config.IgnoreHookErrors,
		"ignore-hook-errors",
		false,
		"ignores errors coming out of hooks",
	)

	flagSet.StringVar(&config.Hooks, "hooks", "", "")

	flagSet.Var(&config.Description, "comment", "Comment for inventory list")
}

func Default() (c Blob) {
	c.PrintOptions = options_print.Default().GetPrintOptions()

	return
}

func (config *Blob) ApplyPrintOptionsConfig(
	printOptions options_print.Options,
) {
	cliSet := config.PrintOptions
	config.PrintOptions = printOptions
	config.PrintOptions.Merge(cliSet, config.maskPrintOptions)
}

func (config Blob) UsePredictableZettelIds() bool {
	return config.PredictableZettelIds
}

func (config Blob) UsePrintTime() bool {
	return config.PrintOptions.PrintTime
}

func (config Blob) UsePrintTags() bool {
	return config.PrintOptions.PrintTagsAlways
}

func (config Blob) IsDryRun() bool {
	return config.dryRun
}

func (config *Blob) SetDryRun(v bool) {
	config.dryRun = v
}
