package repo_config_cli

import (
	"flag"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/bravo/options_tools"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/delta/debug"
	"code.linenisgreat.com/dodder/go/src/echo/descriptions"
)

type Config struct {
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

	PrintOptions, maskPrintOptions options_print.V0
	ToolOptions                    options_tools.Options

	descriptions.Description
}

// TODO add support for all flags
func (c Config) GetCLIFlags() (flags []string) {
	flags = append(flags, fmt.Sprintf("-print-time=%t", c.PrintOptions.PrintTime))
	flags = append(flags, fmt.Sprintf("-print-colors=%t", c.PrintOptions.PrintColors))

	if c.Verbose {
		flags = append(flags, "-verbose")
	}

	return
}

func (c *Config) SetFlagSet(flagSet *flag.FlagSet) {
	flagSet.StringVar(&c.BasePath, "dir-dodder", "", "")

	flagSet.Var(&c.Debug, "debug", "debugging options")
	flagSet.BoolVar(&c.Todo, "todo", false, "")
	flagSet.BoolVar(&c.dryRun, "dry-run", false, "")
	flagSet.BoolVar(&c.Verbose, "verbose", false, "")
	flagSet.BoolVar(&c.Quiet, "quiet", false, "")

	flagSet.BoolVar(&c.CheckoutCacheEnabled, "checkout-cache-enabled", false, "")

	flagSet.BoolVar(
		&c.PredictableZettelIds,
		"predictable-zettel-ids",
		false,
		"generate new zettel ids in order",
	)

	c.PrintOptions.AddToFlags(flagSet, &c.maskPrintOptions)
	c.ToolOptions.AddToFlags(flagSet)

	flagSet.BoolVar(
		&c.PrintOptions.Newlines,
		"zittish-newlines",
		false,
		"add extra newlines to zittish to improve readability",
	)

	flagSet.BoolVar(
		&c.IgnoreHookErrors,
		"ignore-hook-errors",
		false,
		"ignores errors coming out of hooks",
	)

	flagSet.StringVar(&c.Hooks, "hooks", "", "")

	flagSet.Var(&c.Description, "comment", "Comment for inventory list")
}

func Default() (c Config) {
	c.PrintOptions = options_print.Default()

	return
}

func (c *Config) ApplyPrintOptionsConfig(
	po options_print.V0,
) {
	cliSet := c.PrintOptions
	c.PrintOptions = po
	c.PrintOptions.Merge(cliSet, c.maskPrintOptions)
}

func (c Config) UsePredictableZettelIds() bool {
	return c.PredictableZettelIds
}

func (c Config) UsePrintTime() bool {
	return c.PrintOptions.PrintTime
}

func (c Config) UsePrintTags() bool {
	return c.PrintOptions.PrintTagsAlways
}

func (c Config) IsDryRun() bool {
	return c.dryRun
}

func (c *Config) SetDryRun(v bool) {
	c.dryRun = v
}
