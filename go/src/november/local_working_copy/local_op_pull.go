package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
)

func (local *Repo) PullQueryGroupFromRemote(
	remote repo.Repo,
	qg *query.Query,
	options repo.RemoteTransferOptions,
) (err error) {
	return local.pullQueryGroupFromWorkingCopy(
		remote.(repo.WorkingCopy),
		qg,
		options,
	)
}

func (local *Repo) pullQueryGroupFromWorkingCopy(
	remote repo.WorkingCopy,
	queryGroup *query.Query,
	options repo.RemoteTransferOptions,
) (err error) {
	var list *sku.List

	if list, err = remote.MakeInventoryList(queryGroup); err != nil {
		err = errors.Wrap(err)
		return
	}

	importerOptions := sku.ImporterOptions{
		CheckedOutPrinter:   local.PrinterCheckedOutConflictsForRemoteTransfers(),
		AllowMergeConflicts: options.AllowMergeConflicts,
		BlobGenres:          options.BlobGenres,
		ExcludeObjects:      !options.IncludeObjects,
	}

	if options.IncludeBlobs {
		importerOptions.RemoteBlobStore = remote.GetBlobStore()
	}

	importerOptions.ParentNegotiator = ParentNegotiatorFirstAncestor{
		Local:  local,
		Remote: remote,
	}

	importerOptions.PrintCopies = options.PrintCopies
	importer := local.MakeImporter(
		importerOptions,
		sku.GetStoreOptionsImport(),
	)

	if err = local.ImportList(
		list,
		importer,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
