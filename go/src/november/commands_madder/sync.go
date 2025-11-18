package commands_madder

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/india/env_dir"
	"code.linenisgreat.com/dodder/go/src/kilo/command"
	"code.linenisgreat.com/dodder/go/src/lima/command_components_madder"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
	"code.linenisgreat.com/dodder/go/src/mike/blob_transfers"
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
	envBlobStore := cmd.MakeEnvBlobStore(req)
	blobStores := cmd.MakeBlobStoresFromIdsOrAll(req, envBlobStore)

	// TODO output TAP
	ui.Out().Print("Blob Stores:")

	if len(blobStores) == 1 {
		errors.ContextCancelWithBadRequestf(
			req,
			"only one blob store, nothing to sync",
		)

		return
	}

	primary, blobStores := envBlobStore.GetDefaultBlobStoreAndRemaining()

	blobImporter := blob_transfers.MakeBlobImporter(
		envBlobStore,
		primary,
		blobStores,
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
