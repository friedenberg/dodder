package env_repo

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blob_store_id"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

// Config used to initialize a repo for the first time
type BigBang struct {
	GenesisConfig        *genesis_configs.TypedConfigPrivateMutable
	TypedBlobStoreConfig *blob_store_configs.TypedMutableConfig

	InventoryListType ids.Type

	Yin                  string
	Yang                 string
	ExcludeDefaultType   bool
	ExcludeDefaultConfig bool
	OverrideXDGWithCwd   bool
	BlobStoreId          blob_store_id.Id
}

var _ interfaces.CommandComponentWriter = (*BigBang)(nil)

func (bigBang *BigBang) SetDefaults() {
	bigBang.GenesisConfig = genesis_configs.Default()
	bigBang.InventoryListType = ids.GetOrPanic(
		ids.TypeInventoryListVCurrent,
	).Type

	if !store_version.IsCurrentVersionLessOrEqualToV10() {
		bigBang.TypedBlobStoreConfig = blob_store_configs.Default()
	}
}

// TODO switch to flagset wrapper that enforces non-empty descriptions
func (bigBang *BigBang) SetFlagDefinitions(
	flagSet interfaces.CLIFlagDefinitions,
) {
	flagSet.Var(
		&bigBang.InventoryListType,
		"inventory_list-type",
		"the type that will be used when creating inventory lists for this repo",
	)

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

	bigBang.GenesisConfig.Blob.SetFlagDefinitions(flagSet)

	if !store_version.IsCurrentVersionLessOrEqualToV10() {
		bigBang.TypedBlobStoreConfig.Blob.SetFlagDefinitions(flagSet)
	}

	flagSet.Var(
		&bigBang.BlobStoreId,
		"blob_store-id",
		"The name of the existing madder blob store to use",
	)
}
