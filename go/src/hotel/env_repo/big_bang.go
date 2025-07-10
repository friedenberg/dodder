package env_repo

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/delta/genesis_config"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_config"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
)

// Config used to initialize a repo for the first time
type BigBang struct {
	ids.Type
	GenesisConfig   *genesis_config.Current // TODO determine why this is a pointer
	BlobStoreConfig blob_store_config.Current

	Yin                  string
	Yang                 string
	ExcludeDefaultType   bool
	ExcludeDefaultConfig bool
	OverrideXDGWithCwd   bool
}

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

	bigBang.Type = builtin_types.GetOrPanic(
		builtin_types.ImmutableConfigV1,
	).Type

	bigBang.GenesisConfig = genesis_config.Default()
	bigBang.GenesisConfig.SetFlagSet(flagSet)

	// bigBang.BlobStoreConfig = blob_store_config.Default()
	// bigBang.BlobStoreConfig.SetFlagSet(flagSet)
}
