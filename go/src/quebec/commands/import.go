package commands

import (
	"flag"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/inventory_list_coders"
	"code.linenisgreat.com/dodder/go/src/mike/importer"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register(
		"import",
		&Import{
			Version: store_version.VCurrent,
		},
	)
}

// Switch to External store
type Import struct {
	command_components.LocalWorkingCopy
	command_components.RemoteBlobStore

	store_version.Version
	InventoryList string
	PrintCopies   bool
	sku.Proto
}

func (cmd *Import) SetFlagSet(f *flag.FlagSet) {
	f.Var(&cmd.Version, "store-version", "")
	f.StringVar(&cmd.InventoryList, "inventory-list", "", "")
	cmd.RemoteBlobStore.SetFlagSet(f)
	f.BoolVar(
		&cmd.PrintCopies,
		"print-copies",
		true,
		"output when blobs are copied",
	)

	cmd.Proto.SetFlagSet(f)
}

func (cmd Import) Run(req command.Request) {
	localWorkingCopy := cmd.MakeLocalWorkingCopy(req)

	if cmd.InventoryList == "" {
		errors.ContextCancelWithBadRequestf(req, "empty inventory list")
	}

	bf := localWorkingCopy.GetStore().GetInventoryListStore().FormatForVersion(
		cmd.Version,
	)

	var readCloser io.ReadCloser

	// setup inventory list reader
	{

		var err error

		if readCloser, err = env_dir.NewFileReader(
			env_dir.MakeConfig(
				// TODO ready from config
				sha.Env{},
				env_dir.MakeHashBucketPathJoinFunc(cmd.Config.GetHashBuckets()),
				cmd.Config.GetBlobCompression(),
				cmd.Config.GetBlobEncryption(),
			),
			cmd.InventoryList,
		); err != nil {
			localWorkingCopy.Cancel(err)
		}

		defer errors.ContextMustClose(localWorkingCopy, readCloser)
	}

	bufferedReader, repoolBufferedReader := pool.GetBufferedReader(readCloser)
	defer repoolBufferedReader()

	list := sku.MakeList()

	// TODO determine why this is not erroring for invalid input
	if err := inventory_list_coders.CollectSkuList(
		req,
		bf,
		bufferedReader,
		list,
	); err != nil {
		localWorkingCopy.Cancel(err)
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
				localWorkingCopy.Cancel(err)
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

		localWorkingCopy.Cancel(err)
	}
}
