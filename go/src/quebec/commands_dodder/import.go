package commands_dodder

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/india/command_components_madder"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/blob_transfers"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
	"code.linenisgreat.com/dodder/go/src/mike/remote_transfer"
	"code.linenisgreat.com/dodder/go/src/papa/command_components_dodder"
)

func init() {
	utility.AddCmd("import", &Import{})
}

type Import struct {
	command_components_dodder.LocalWorkingCopy
	command_components_madder.BlobStore

	repo.ImporterOptions

	sku.Proto
}

var _ interfaces.CommandComponentWriter = (*Import)(nil)

func (cmd *Import) SetFlagDefinitions(
	flagDefinitions interfaces.CLIFlagDefinitions,
) {
	cmd.ImporterOptions.SetFlagDefinitions(flagDefinitions)
	cmd.Proto.SetFlagDefinitions(flagDefinitions)
}

func (cmd Import) Run(req command.Request) {
	inventoryListPath := req.PopArg("inventory_list-path")
	blobStoreBasePath := req.PopArg("blob_store-base-path")
	blobStoreConfigPath := req.PopArg("blob_store-config-path")

	local := cmd.MakeLocalWorkingCopy(req)

	if inventoryListPath == "" {
		errors.ContextCancelWithBadRequestf(req, "empty inventory list")
	}

	var readCloser io.ReadCloser

	// setup inventory list reader
	{
		var err error

		if readCloser, err = files.Open(
			inventoryListPath,
		); err != nil {
			local.Cancel(err)
		}

		defer errors.ContextMustClose(local, readCloser)
	}

	bufferedReader, repoolBufferedReader := pool.GetBufferedReader(readCloser)
	defer repoolBufferedReader()

	inventoryListCoderCloset := local.GetInventoryListCoderCloset()

	cmd.DedupingFormatId = markl.PurposeV5MetadataDigestWithoutTai
	cmd.CheckedOutPrinter = local.PrinterCheckedOutConflictsForRemoteTransfers()

	if blobStoreConfigPath != "" {
		cmd.RemoteBlobStore = cmd.MakeBlobStoreFromConfigPath(
			local.GetEnvRepo().GetEnvBlobStore(),
			blobStoreBasePath,
			blobStoreConfigPath,
		)

		if cmd.RemoteBlobStore.GetBlobStore() != nil &&
			cmd.RemoteBlobStore.Path.GetBase() == "" {
			req.Cancel(errors.Errorf("missing blob store base path"))
			return
		}
	}

	var afterDecoding func(*sku.Transacted) error

	blobImporter := blob_transfers.MakeBlobImporter(
		local.GetEnvRepo().GetEnvBlobStore(),
		cmd.RemoteBlobStore,
		local.GetBlobStore(),
	)

	blobImporter.UseDestinationHashType = true

	blobImporter.CopierDelegate = sku.MakeBlobCopierDelegate(
		local.GetUI(),
		false,
	)

	// TODO traverse object graph and rewrite all signature in topological order
	// TODO move this to the importer directly
	if cmd.OverwriteSignatures {
		afterDecoding = func(object *sku.Transacted) (err error) {
			object.Metadata.GetObjectDigestMutable().Reset()
			object.Metadata.GetObjectSigMutable().Reset()
			object.Metadata.GetRepoPubKeyMutable().Reset()

			if err = blobImporter.ImportBlobIfNecessary(
				object.Metadata.GetBlobDigest(),
				object,
			); err != nil { // TODO rewrite blob
				var errNotEqual markl.ErrNotEqual

				if errors.As(err, &errNotEqual) {
					if errNotEqual.IsDifferentHashTypes() {
						err = nil
						object.Metadata.GetBlobDigestMutable().ResetWithMarklId(
							errNotEqual.Actual,
						)
					} else {
						err = errors.Wrap(err)
						return err
					}
				} else if env_dir.IsErrBlobAlreadyExists(err) {
					err = nil
				} else {
					err = errors.Wrap(err)
					return err
				}
			}

			// TODO add mother?
			// TODO rewrite time?
			if err = object.FinalizeAndSignOverwrite(
				local.GetEnvRepo().GetConfigPrivate().Blob,
			); err != nil {
				err = errors.Wrap(err)
				return err
			}

			return err
		}
	}

	seq := inventoryListCoderCloset.AllDecodedObjectsFromStream(
		bufferedReader,
		afterDecoding,
	)

	importer := local.MakeImporter(
		cmd.ImporterOptions,
		sku.GetStoreOptionsImport(),
	)

	if err := local.ImportSeq(
		seq,
		importer,
	); err != nil {
		if !errors.Is(err, remote_transfer.ErrNeedsMerge) {
			err = errors.Wrap(err)
		}

		local.Cancel(err)
	}
}
