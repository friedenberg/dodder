package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/queries"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
	"code.linenisgreat.com/dodder/go/src/mike/remote_transfer"
)

func (local *Repo) PullQueryGroupFromRemote(
	remote repo.Repo,
	qg *queries.Query,
	options repo.ImporterOptions,
) (err error) {
	if err = local.pullQueryGroupFromWorkingCopy(
		remote,
		qg,
		options,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (local *Repo) pullQueryGroupFromWorkingCopy(
	remote repo.Repo,
	queryGroup *queries.Query,
	importerOptions repo.ImporterOptions,
) (err error) {
	var list *sku.ListTransacted

	if list, err = remote.MakeInventoryList(queryGroup); err != nil {
		err = errors.Wrap(err)
		return err
	}

	importerOptions.CheckedOutPrinter = local.PrinterCheckedOutConflictsForRemoteTransfers()

	if !importerOptions.ExcludeBlobs {
		remoteBlobStore := remote.GetBlobStore()
		importerOptions.RemoteBlobStore = remoteBlobStore
	}

	importerOptions.ParentNegotiator = ParentNegotiatorFirstAncestor{
		Local:  local,
		Remote: remote,
	}

	importer := local.MakeImporter(
		importerOptions,
		sku.GetStoreOptionsImport(),
	)

	if err = local.ImportSeq(
		quiter.MakeSeqErrorFromSeq(list.All()),
		importer,
	); err != nil {
		if errors.Is(err, remote_transfer.ErrNeedsMerge) {
			err = errors.WithoutStack(err)
		} else {
			err = errors.Wrap(err)
		}

		return err
	}

	return err
}
