package commands

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/mike/importer"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("import", &Import{})
}

// Switch to External store
type Import struct {
	command_components.LocalWorkingCopy
	command_components.RemoteBlobStore

	InventoryList       string
	PrintCopies         bool
	OverwriteSignatures bool

	sku.Proto
}

func (cmd *Import) SetFlagSet(flagSet interfaces.CommandLineFlagDefinitions) {
	flagSet.StringVar(&cmd.InventoryList, "inventory-list", "", "")
	cmd.RemoteBlobStore.SetFlagSet(flagSet)

	flagSet.BoolVar(
		&cmd.PrintCopies,
		"print-copies",
		true,
		"output when blobs are copied",
	)

	flagSet.BoolVar(
		&cmd.OverwriteSignatures,
		"overwrite-signatures",
		false,
		"ignore object pubkeys and signatures and generate new ones (causing this repo to create the objects as new instead of importing them)",
	)

	cmd.Proto.SetFlagSet(flagSet)
}

func (cmd Import) Run(req command.Request) {
	repo := cmd.MakeLocalWorkingCopy(req)

	if cmd.InventoryList == "" {
		errors.ContextCancelWithBadRequestf(req, "empty inventory list")
	}

	var readCloser io.ReadCloser

	// setup inventory list reader
	{
		var err error

		if readCloser, err = env_dir.NewFileReader(
			env_dir.MakeConfig(
				markl.HashTypeSha256,
				env_dir.MakeHashBucketPathJoinFunc(cmd.Config.GetHashBuckets()),
				cmd.Config.GetBlobCompression(),
				cmd.Config.GetBlobEncryption(),
			),
			cmd.InventoryList,
		); err != nil {
			repo.Cancel(err)
		}

		defer errors.ContextMustClose(repo, readCloser)
	}

	bufferedReader, repoolBufferedReader := pool.GetBufferedReader(readCloser)
	defer repoolBufferedReader()

	inventoryListCoderCloset := repo.GetInventoryListCoderCloset()

	var afterDecoding func(*sku.Transacted) error

	if cmd.OverwriteSignatures {
		afterDecoding = func(object *sku.Transacted) (err error) {
			object.Metadata.GetObjectDigestMutable().Reset()
			object.Metadata.GetObjectSigMutable().Reset()
			object.Metadata.GetRepoPubKeyMutable().Reset()

			if err = object.FinalizeAndSignOverwrite(
				repo.GetEnvRepo().GetConfigPrivate().Blob,
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}
	}

	seq := inventoryListCoderCloset.AllDecodedObjectsFromStream(
		bufferedReader,
		afterDecoding,
	)

	importerOptions := sku.ImporterOptions{
		DedupingFormatId:  markl.FormatIdV5MetadataDigestWithoutTai,
		CheckedOutPrinter: repo.PrinterCheckedOutConflictsForRemoteTransfers(),
	}

	if cmd.BasePath != "" {
		{
			var err error

			if importerOptions.RemoteBlobStore, err = cmd.MakeRemoteBlobStore(
				repo,
			); err != nil {
				repo.Cancel(err)
			}
		}
	}

	importerOptions.PrintCopies = cmd.PrintCopies
	importerr := repo.MakeImporter(
		importerOptions,
		sku.GetStoreOptionsImport(),
	)

	if err := repo.ImportSeq(
		seq,
		importerr,
	); err != nil {
		if !errors.Is(err, importer.ErrNeedsMerge) {
			err = errors.Wrap(err)
		}

		repo.Cancel(err)
	}
}
