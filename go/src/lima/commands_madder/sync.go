package commands_madder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/india/command_components_madder"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/blob_transfers"
)

func init() {
	utility.AddCmd("sync", &Sync{})
}

type Sync struct {
	command_components_madder.EnvBlobStore
	command_components_madder.BlobStore

	Limit int
}

var _ interfaces.CommandComponentWriter = (*Sync)(nil)

func (cmd *Sync) SetFlagDefinitions(
	flagSet interfaces.CLIFlagDefinitions,
) {
	flagSet.IntVar(
		&cmd.Limit,
		"limit",
		0,
		"number of blobs to sync before stopping. 0 means don't stop (full consent)",
	)
}

// TODO add completion for blob store id's

func (cmd Sync) Run(req command.Request) {
	// blobStoreIds := req.PopArgs()
	cmd.runAllStores(req)
}

func (cmd Sync) runAllStores(req command.Request) {
	req.AssertNoMoreArgs()
	envBlobStore := cmd.MakeEnvBlobStore(req)
	blobStoresInitialized := envBlobStore.GetBlobStores()
	blobStores := make([]interfaces.BlobStore, len(blobStoresInitialized))

	for idx := range blobStoresInitialized {
		blobStores[idx] = blobStoresInitialized[idx]
	}

	// TODO output TAP
	ui.Out().Print("Blob Stores:")

	for i, blobStore := range blobStoresInitialized {
		ui.Out().Printf("%d: %s", i, blobStore.Name)
	}

	if len(blobStoresInitialized) == 1 {
		errors.ContextCancelWithBadRequestf(
			req,
			"only one blob store, nothing to sync",
		)
		return
	}

	primary := blobStoresInitialized[0]
	blobStores = blobStores[1:]

	blobImporter := blob_transfers.MakeBlobImporter(
		envBlobStore,
		primary,
		blobStores...,
	)

	blobImporter.CopierDelegate = sku.MakeBlobCopierDelegate(
		envBlobStore.GetUI(),
		false,
	)

	defer req.Must(
		func(_ interfaces.ActiveContext) error {
			ui.Err().Printf(
				"Successes: %d, Failures: %d, Ignored: %d, Total: %d",
				blobImporter.Counts.Succeeded,
				blobImporter.Counts.Failed,
				blobImporter.Counts.Ignored,
				blobImporter.Counts.Total,
			)

			return nil
		},
	)

	for blobId, errIter := range primary.AllBlobs() {
		if errIter != nil {
			ui.Err().Print(errIter)
			continue
		}

		if err := blobImporter.ImportBlobIfNecessary(blobId, nil); err != nil {
			var errNotEqual markl.ErrNotEqual

			if errors.As(err, &errNotEqual) {
				ui.Err().Printf(
					"%q -> %q",
					errNotEqual.Expected,
					errNotEqual.Actual,
				)
			} else if !env_dir.IsErrBlobAlreadyExists(err) {
				ui.Err().Print(err)
			}
		}

		if cmd.Limit > 0 &&
			(blobImporter.Counts.Succeeded+blobImporter.Counts.Failed) > cmd.Limit {
			ui.Err().Print("limit hit, stopping")
			break
		}
	}
}
