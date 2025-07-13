package commands

import (
	"flag"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/inventory_list_blobs"
	"code.linenisgreat.com/dodder/go/src/mike/importer"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register(
		"import",
		&Import{
			StoreVersion: store_version.VCurrent,
		},
	)
}

// Switch to External store
type Import struct {
	command_components.LocalWorkingCopy
	command_components.RemoteBlobStore

	genesis_configs.StoreVersion
	InventoryList string
	PrintCopies   bool
	sku.Proto
}

func (cmd *Import) SetFlagSet(f *flag.FlagSet) {
	f.Var(&cmd.StoreVersion, "store-version", "")
	f.StringVar(&cmd.InventoryList, "inventory-list", "", "")
	cmd.RemoteBlobStore.SetFlagSet(f)
	f.BoolVar(&cmd.PrintCopies, "print-copies", true, "output when blobs are copied")

	cmd.Proto.SetFlagSet(f)
}

func (cmd Import) Run(dep command.Request) {
	localWorkingCopy := cmd.MakeLocalWorkingCopy(dep)

	if cmd.InventoryList == "" {
		dep.CancelWithBadRequestf("empty inventory list")
	}

	bf := localWorkingCopy.GetStore().GetInventoryListStore().FormatForVersion(cmd.StoreVersion)

	var readCloser io.ReadCloser

	// setup inventory list reader
	{
		o := env_dir.FileReadOptions{
			Config: env_dir.MakeConfig(
				cmd.Config.GetBlobCompression(),
				cmd.Config.GetBlobEncryption(),
			),
			Path: cmd.InventoryList,
		}

		var err error

		if readCloser, err = env_dir.NewFileReader(o); err != nil {
			localWorkingCopy.CancelWithError(err)
		}

		defer localWorkingCopy.MustClose(readCloser)
	}

	bufferedReader := ohio.BufferedReader(readCloser)
	defer pool.GetBufioReader().Put(bufferedReader)

	list := sku.MakeList()

	// TODO determine why this is not erroring for invalid input
	if err := inventory_list_blobs.ReadInventoryListBlob(
		bf,
		bufferedReader,
		list,
	); err != nil {
		localWorkingCopy.CancelWithError(err)
	}

	importerOptions := sku.ImporterOptions{
		CheckedOutPrinter: localWorkingCopy.PrinterCheckedOutConflictsForRemoteTransfers(),
	}

	if cmd.Blobs != "" {
		{
			var err error

			if importerOptions.RemoteBlobStore, err = cmd.MakeRemoteBlobStore(
				localWorkingCopy,
			); err != nil {
				localWorkingCopy.CancelWithError(err)
			}
		}
	}

	importerOptions.PrintCopies = cmd.PrintCopies
	i := localWorkingCopy.MakeImporter(
		importerOptions,
		sku.GetStoreOptionsImport(),
	)

	if err := localWorkingCopy.ImportList(
		list,
		i,
	); err != nil {
		if !errors.Is(err, importer.ErrNeedsMerge) {
			err = errors.Wrap(err)
		}

		localWorkingCopy.CancelWithError(err)
	}
}
