package store_fs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

// TODO what does this even do. This caused [cervicis/marshall.zettel !task pom-2 project-2021-zit-bugs project-25q1-zit_workspaces-crit] fix issue with tags other than workspace in `checkin -organize` bei…
// likely due to this method overriding tags that were set by organize. maybe
// this bug existed before workspaces?
func (store *Store) RefreshCheckedOut(
	co *sku.CheckedOut,
) (err error) {
	var item *sku.FSItem

	if item, err = store.ReadFSItemFromExternal(co.GetSkuExternal()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = store.HydrateExternalFromItem(
		sku.CommitOptions{
			StoreOptions: sku.StoreOptions{
				UpdateTai: true,
			},
		},
		item,
		co.GetSku(),
		co.GetSkuExternal(),
	); err != nil {
		if sku.IsErrMergeConflict(err) {
			co.SetState(checked_out_state.Conflicted)

			if err = co.GetSkuExternal().ObjectId.SetWithIdLike(
				&co.GetSku().ObjectId,
			); err != nil {
				err = errors.Wrap(err)
				return err
			}
		} else {
			err = errors.Wrapf(err, "Cwd: %#v", item.Debug())
			return err
		}
	}

	return err
}

func (store *Store) ReadCheckedOutFromTransacted(
	sk2 *sku.Transacted,
) (co *sku.CheckedOut, err error) {
	co = GetCheckedOutPool().Get()

	if err = store.readIntoCheckedOutFromTransacted(sk2, co); err != nil {
		err = errors.Wrap(err)
		return co, err
	}

	return co, err
}

func (store *Store) readIntoCheckedOutFromTransacted(
	sk *sku.Transacted,
	co *sku.CheckedOut,
) (err error) {
	if co.GetSku() != sk {
		sku.Resetter.ResetWith(co.GetSku(), sk)
	}

	ok := false

	var kfp *sku.FSItem

	if kfp, ok = store.Get(&sk.ObjectId); !ok {
		err = collections.MakeErrNotFound(sk.GetObjectId())
		return err
	}

	if err = store.HydrateExternalFromItem(
		sku.CommitOptions{
			StoreOptions: sku.StoreOptions{
				UpdateTai: true,
			},
		},
		kfp,
		sk,
		co.GetSkuExternal(),
	); err != nil {
		if errors.IsNotExist(err) {
			err = errors.MakeErrStopIteration()
		} else if sku.IsErrMergeConflict(err) {
			co.SetState(checked_out_state.Conflicted)

			if err = co.GetSkuExternal().ObjectId.SetWithIdLike(
				&sk.ObjectId,
			); err != nil {
				err = errors.Wrap(err)
				return err
			}
		} else {
			err = errors.Wrapf(err, "Cwd: %#v", kfp)
		}

		return err
	}

	return err
}
