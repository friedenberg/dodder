package env_repo

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
)

// Config used to initialize a repo for the first time
type BigBang struct {
	GenesisConfig        *genesis_configs.TypedBlobPrivateMutable
	TypedBlobStoreConfig *blob_store_configs.TypedMutableConfig

	Yin                  string
	Yang                 string
	ExcludeDefaultType   bool
	ExcludeDefaultConfig bool
	OverrideXDGWithCwd   bool
}

func (bigBang *BigBang) SetDefaults() {
	bigBang.GenesisConfig = genesis_configs.Default()

	if !store_version.IsCurrentVersionLessOrEqualToV10() {
		bigBang.TypedBlobStoreConfig = blob_store_configs.Default()
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

	bigBang.GenesisConfig.Blob.SetFlagSet(flagSet)

	if !store_version.IsCurrentVersionLessOrEqualToV10() {
		bigBang.TypedBlobStoreConfig.Blob.SetFlagSet(flagSet)
	}
}
