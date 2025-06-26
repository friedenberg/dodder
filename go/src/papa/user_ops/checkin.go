package user_ops

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/lima/organize_text"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

type Checkin struct {
	Proto sku.Proto

	// TODO make flag family disambiguate these options
	// and use with other commands too
	Delete             bool
	RefreshCheckout    bool
	Organize           bool
	CheckoutBlobAndRun string
	OpenBlob           bool
	Edit               bool // TODO add support back for this
}

func (op Checkin) Run(
	repo *local_working_copy.Repo,
	q *query.Query,
) (err error) {
	var l sync.Mutex

	results := sku.MakeSkuTypeSetMutable()

	if err = repo.GetStore().QuerySkuType(
		q,
		func(co sku.SkuType) (err error) {
			l.Lock()
			defer l.Unlock()

			return results.Add(co.Clone())
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if op.Organize {
		if err = op.runOrganize(repo, q, results); err != nil {
			err = errors.Wrap(err)
			return
		}

		object_metadata.Resetter.Reset(&op.Proto.Metadata)
	}

	var processed sku.TransactedMutableSet

	if processed, err = repo.Checkin(
		results,
		op.Proto,
		op.Delete,
		op.RefreshCheckout,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = op.openBlobIfNecessary(repo, processed); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (op Checkin) runOrganize(
	repo *local_working_copy.Repo,
	query *query.Query,
	results sku.SkuTypeSetMutable,
) (err error) {
	flagDelete := organize_text.OptionCommentBooleanFlag{
		Value:   &op.Delete,
		Comment: "delete once checked in",
	}

	opOrganize := Organize2{
		Repo: repo,
		Metadata: organize_text.Metadata{
			TagSet: op.Proto.Tags,
			Type:   op.Proto.Type,
			RepoId: query.RepoId,
			OptionCommentSet: organize_text.MakeOptionCommentSet(
				map[string]organize_text.OptionComment{
					"delete": flagDelete,
				},
				&organize_text.OptionCommentUnknown{
					Value: "instructions: to prevent an object from being checked in, delete it entirely",
				},
				organize_text.OptionCommentWithKey{
					Key:           "delete",
					OptionComment: flagDelete,
				},
			),
		},
	}

	ui.Log().Print(query)

	var organizeResults organize_text.OrganizeResults

	if organizeResults, err = opOrganize.Run(results); err != nil {
		err = errors.Wrap(err)
		return
	}

	var changes organize_text.Changes

	if changes, err = organize_text.ChangesFromResults(
		repo.GetConfig().GetCLIConfig().PrintOptions,
		organizeResults,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, co := range changes.After.AllSkuAndIndex() {
		if err = results.Add(co.Clone()); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	for _, co := range changes.Removed.AllSkuAndIndex() {
		if err = results.Del(co); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (c Checkin) openBlobIfNecessary(
	repo *local_working_copy.Repo,
	objects sku.TransactedSet,
) (err error) {
	if !c.OpenBlob && c.CheckoutBlobAndRun == "" {
		return
	}

	opCheckout := Checkout{
		Repo: repo,
		Options: checkout_options.Options{
			CheckoutMode: checkout_mode.BlobOnly,
		},
		Utility: c.CheckoutBlobAndRun,
	}

	if _, err = opCheckout.Run(objects); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
