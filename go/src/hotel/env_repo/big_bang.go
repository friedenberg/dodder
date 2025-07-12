package env_repo

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_config"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_config"
)

// Config used to initialize a repo for the first time
type BigBang struct {
	GenesisConfig   genesis_config.PrivateMutable
	BlobStoreConfig blob_store_config.ConfigMutable

	Yin                  string
	Yang                 string
	ExcludeDefaultType   bool
	ExcludeDefaultConfig bool
	OverrideXDGWithCwd   bool
}

func (bigBang *BigBang) SetDefaults() {
	bigBang.GenesisConfig = genesis_config.DefaultMutable()

	if !store_version.IsCurrentVersionLessOrEqualToV10() {
		bigBang.BlobStoreConfig = blob_store_config.Default()
	}
}

// TODO switch to flagset wrapper that enforces non-empty descriptions
func (bigBang *BigBang) SetFlagSet(flagSet *flag.FlagSet) {
	flagSet.BoolVar(
		&bigBang.OverrideXDGWithCwd,
		"override-xdg-with-cwd",
		false,
		"don't use XDG for this repo, and instead use the CWD and make a `.dodder` directory",
	)

	flagSet.StringVar(
		&bigBang.Yin,
		"yin",
		"",
		"File containing list of zettel id left parts",
	)

	flagSet.StringVar(
		&bigBang.Yang,
		"yang",
		"",
		"File containing list of zettel id right parts",
	)

	bigBang.SetDefaults()

	bigBang.GenesisConfig.SetFlagSet(flagSet)

	if !store_version.IsCurrentVersionLessOrEqualToV10() {
		bigBang.BlobStoreConfig.SetFlagSet(flagSet)
	}
}
