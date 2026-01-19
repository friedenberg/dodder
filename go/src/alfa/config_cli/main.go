package config_cli

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/cli"
	"code.linenisgreat.com/dodder/go/src/delta/debug"
)

type Config struct {
	Debug   debug.Options
	Verbose bool
	Quiet   bool
	Todo    bool
	dryRun  bool
}

var _ interfaces.CommandComponentWriter = (*Config)(nil)

func (config *Config) SetFlagDefinitions(flagSet interfaces.CLIFlagDefinitions) {
	cli.FlagSetVarWithCompletion(flagSet, &config.Debug, "debug")
	flagSet.BoolVar(&config.Todo, "todo", false, "")
	flagSet.BoolVar(&config.dryRun, "dry-run", false, "")
	flagSet.BoolVar(&config.Verbose, "verbose", false, "")
	flagSet.BoolVar(&config.Quiet, "quiet", false, "")
}

func Default() (config Config) {
	return config
}

func (config Config) IsDryRun() bool {
	return config.dryRun
}

func (config *Config) SetDryRun(v bool) {
	config.dryRun = v
}
