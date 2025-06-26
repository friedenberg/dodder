package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/env_workspace"
)

func (repo *Repo) Checkin(
	skus sku.SkuTypeSetMutable,
	proto sku.Proto,
	delete bool,
	refreshCheckout bool,
) (processed sku.TransactedMutableSet, err error) {
	repo.Must(repo.Lock)

	processed = sku.MakeTransactedMutableSet()
	sortedResults := quiter.ElementsSorted(
		skus,
		func(left, right sku.SkuType) bool {
			return left.String() < right.String()
		},
	)

	for _, co := range sortedResults {
		if refreshCheckout {
			if err = repo.GetEnvWorkspace().GetStoreFS().RefreshCheckedOut(
				co,
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		external := co.GetSkuExternal()

		if co.GetState() == checked_out_state.Untracked &&
			(co.GetSkuExternal().GetGenre() == genres.Zettel ||
				co.GetSkuExternal().GetGenre() == genres.Blob) {
			if external.Metadata.IsEmpty() {
				continue
			}

			if err = repo.GetStore().UpdateTransactedFromBlobs(
				co,
			); err != nil {
				if errors.Is(err, env_workspace.ErrUnsupportedOperation{}) {
					err = nil
				} else {
					err = errors.Wrap(err)
					return
				}
			}

			external.ObjectId.Reset()

			proto.Apply(external, genres.Zettel)

			if err = repo.GetStore().CreateOrUpdate(
				external,
				sku.CommitOptions{
					Proto: proto,
				},
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		} else {
			if err = repo.GetStore().CreateOrUpdateCheckedOut(
				co,
				!delete,
			); err != nil {
				err = errors.Wrapf(err, "CheckedOut: %s", co)
				return
			}
		}

		if !delete {
			continue
		}

		if err = repo.GetStore().DeleteCheckedOut(co); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = processed.Add(co.GetSkuExternal().CloneTransacted()); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	repo.Must(repo.Unlock)

	return
}
