package commands

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
	"code.linenisgreat.com/dodder/go/src/lima/repo"
	"code.linenisgreat.com/dodder/go/src/mike/remote_transfer"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("import", &Import{})
}

type Import struct {
	command_components.LocalWorkingCopy
	command_components_madder.BlobStore

	repo.ImporterOptions

	sku.Proto
}

var _ interfaces.CommandComponentWriter = (*Import)(nil)

func (cmd *Import) SetFlagDefinitions(
	flagDefinitions interfaces.CommandLineFlagDefinitions,
) {
	cmd.ImporterOptions.SetFlagDefinitions(flagDefinitions)
	cmd.Proto.SetFlagDefinitions(flagDefinitions)
}

func (cmd Import) Run(req command.Request) {
	inventoryListPath := req.PopArg("inventory_list-path")
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
		cmd.RemoteBlobStore = cmd.MakeBlobStore(
			local.GetEnvRepo(),
			blobStoreConfigPath,
		)
	}

	var afterDecoding func(*sku.Transacted) error

	blobImporter := remote_transfer.MakeBlobImporter(
		local.GetEnvRepo(),
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
						return
					}
				} else if env_dir.IsErrBlobAlreadyExists(err) {
					err = nil
				} else {
					err = errors.Wrap(err)
					return
				}
			}

			// TODO add mother?
			// TODO rewrite time?
			if err = object.FinalizeAndSignOverwrite(
				local.GetEnvRepo().GetConfigPrivate().Blob,
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
