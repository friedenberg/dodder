package user_ops

import (
	"io"
	"os"
	"sync"

	"code.linenisgreat.com/dodder/go/src/_/vim_cli_options_builder"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_ui"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
	"code.linenisgreat.com/dodder/go/src/oscar/queries"
	"code.linenisgreat.com/dodder/go/src/papa/organize_text"
	"code.linenisgreat.com/dodder/go/src/victor/local_working_copy"
)

// TODO migrate over to Organize2
type Organize struct {
	*local_working_copy.Repo
	organize_text.Metadata
	DontUseQueryGroupForOrganizeMetadata bool
}

func (op Organize) RunWithQueryGroup(
	qg *queries.Query,
) (organizeResults organize_text.OrganizeResults, err error) {
	skus := sku.MakeSkuTypeSetMutable()
	var l sync.RWMutex

	if err = op.GetStore().QueryTransactedAsSkuType(
		qg,
		func(co sku.SkuType) (err error) {
			l.Lock()
			defer l.Unlock()

			return skus.Add(co.Clone())
		},
	); err != nil {
		err = errors.Wrap(err)
		return organizeResults, err
	}

	if organizeResults, err = op.RunWithSkuType(qg, skus); err != nil {
		err = errors.Wrap(err)
		return organizeResults, err
	}

	return organizeResults, err
}

// TODO remove
func (op Organize) RunWithTransacted(
	qg *queries.Query,
	transacted sku.TransactedSet,
) (organizeResults organize_text.OrganizeResults, err error) {
	skus := sku.MakeSkuTypeSetMutable()

	for z := range transacted.All() {
		clone := sku.CloneSkuTypeFromTransacted(
			z.GetSku(),
			checked_out_state.Internal,
		)

		skus.Add(clone)
	}

	if organizeResults, err = op.RunWithSkuType(qg, skus); err != nil {
		err = errors.Wrap(err)
		return organizeResults, err
	}

	return organizeResults, err
}

func (op Organize) RunWithSkuType(
	q *queries.Query,
	skus sku.SkuTypeSet,
) (organizeResults organize_text.OrganizeResults, err error) {
	organizeResults.Original = skus
	organizeResults.QueryGroup = q

	var repoId ids.RepoId

	if q != nil {
		repoId = q.RepoId
	}

	if organizeResults.QueryGroup == nil ||
		op.DontUseQueryGroupForOrganizeMetadata {
		b := op.MakeQueryBuilder(
			ids.MakeGenre(genres.All()...),
			nil,
		).WithExternalLike(
			skus,
		)

		if organizeResults.QueryGroup, err = b.BuildQueryGroup(); err != nil {
			err = errors.Wrap(err)
			return organizeResults, err
		}
	}

	organizeResults.QueryGroup.RepoId = repoId

	organizeFlags := organize_text.MakeFlagsWithMetadata(op.Metadata)
	op.ApplyToOrganizeOptions(&organizeFlags.Options)
	organizeFlags.Skus = skus

	createOrganizeFileOp := CreateOrganizeFile{
		Repo: op.Repo,
		Options: op.Repo.MakeOrganizeOptionsWithQueryGroup(
			organizeFlags,
			organizeResults.QueryGroup,
		),
	}

	typen := queries.GetTypes(organizeResults.QueryGroup)

	if typen.Len() == 1 {
		createOrganizeFileOp.Type = typen.Any()
	}

	var f *os.File

	if f, err = op.GetEnvRepo().GetTempLocal().FileTempWithTemplate(
		"*." + op.GetConfig().GetFileExtensions().Organize,
	); err != nil {
		err = errors.Wrap(err)
		return organizeResults, err
	}

	defer errors.DeferredCloser(&err, f)

	if organizeResults.Before, err = createOrganizeFileOp.RunAndWrite(
		f,
	); err != nil {
		err = errors.Wrap(err)
		return organizeResults, err
	}

	// TODO refactor into common vim processing loop
	for {
		openVimOp := OpenEditor{
			VimOptions: vim_cli_options_builder.New().
				WithFileType("dodder-organize").
				Build(),
		}

		if err = openVimOp.Run(op.Repo, f.Name()); err != nil {
			err = errors.Wrap(err)
			return organizeResults, err
		}

		// if err = op.Reset(); err != nil {
		// 	err = errors.Wrap(err)
		// 	return
		// }

		readOrganizeTextOp := ReadOrganizeFile{}

		if _, err = f.Seek(0, io.SeekStart); err != nil {
			err = errors.Wrap(err)
			return organizeResults, err
		}

		if organizeResults.After, err = readOrganizeTextOp.Run(
			op.Repo,
			f,
			organize_text.NewMetadataWithOptionCommentLookup(
				organizeResults.Before.GetRepoId(),
				op.GetPrototypeOptionComments(),
			),
		); err != nil {
			if op.handleReadChangesError(op.Repo, err) {
				err = nil
				continue
			} else {
				ui.Err().Printf("aborting organize")
				return organizeResults, err
			}
		}

		break
	}

	return organizeResults, err
}

func (cmd Organize) handleReadChangesError(
	envUI env_ui.Env,
	err error,
) (tryAgain bool) {
	var errorRead organize_text.ErrorRead

	if err != nil && !errors.As(err, &errorRead) {
		ui.Err().Printf("unrecoverable organize read failure: %s", err)
		tryAgain = false
		return tryAgain
	}

	return envUI.Retry(
		"reading changes failed",
		"edit and try again?",
		err,
	)
}
