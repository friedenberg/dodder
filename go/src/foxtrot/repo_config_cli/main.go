package repo_config_cli

import (
	"code.linenisgreat.com/dodder/go/src/alfa/cli"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
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
	IgnoreWorkspace  bool

	CheckoutCacheEnabled bool
	PredictableZettelIds bool

	printOptionsOverlay options_print.Overlay
	ToolOptions         options_tools.Options

	descriptions.Description
}

func (config Config) GetPrintOptionsOverlay() options_print.Overlay {
	return config.printOptionsOverlay
}

// TODO add support for all flags
// TODO move to store_config
func (config *Config) SetFlagSet(flagSet *flags.FlagSet) {
	flagSet.StringVar(&config.BasePath, "dir-dodder", "", "")

	cli.FlagSetVarWithCompletion(
		flagSet,
		&config.Debug,
		"debug",
	)

	flagSet.BoolVar(&config.Todo, "todo", false, "")
	flagSet.BoolVar(&config.dryRun, "dry-run", false, "")
	flagSet.BoolVar(&config.Verbose, "verbose", false, "")
	flagSet.BoolVar(&config.Quiet, "quiet", false, "")

	flagSet.BoolVar(
		&config.IgnoreWorkspace,
		"ignore-workspace",
		false,
		"ignore any workspaces that may be present and checkout the object in a temporary workspace",
	)

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

	config.printOptionsOverlay.AddToFlags(flagSet)
	config.ToolOptions.SetFlagSet(flagSet)

	flagSet.BoolVar(
		&config.IgnoreHookErrors,
		"ignore-hook-errors",
		false,
		"ignores errors coming out of hooks",
	)

	flagSet.StringVar(&config.Hooks, "hooks", "", "")

	flagSet.Var(&config.Description, "comment", "Comment for inventory list")
}

func Default() (config Config) {
	// config.printOptionsOverlay =
	// options_print.DefaultOverlay().GetPrintOptionsOverlay()

	return
}

// func (config Config) GetPrintOptions() options_print.Options {
// 	return options_print.MakeDefaultConfig(config.printOptionsOverlay)
// }

func (config Config) UsePredictableZettelIds() bool {
	return config.PredictableZettelIds
}

func (config Config) IsDryRun() bool {
	return config.dryRun
}

func (config *Config) SetDryRun(v bool) {
	config.dryRun = v
}
