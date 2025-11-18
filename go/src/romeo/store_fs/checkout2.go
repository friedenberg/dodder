package store_fs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

func (store *Store) checkoutOneIfNecessary(
	options checkout_options.Options,
	tg sku.TransactedGetter,
) (co *sku.CheckedOut, item *sku.FSItem, err error) {
	internal := tg.GetSku()
	co = GetCheckedOutPool().Get()

	sku.Resetter.ResetWith(co.GetSku(), internal)

	var alreadyCheckedOut bool

	if item, alreadyCheckedOut, err = store.prepareFSItemForCheckOut(options, co); err != nil {
		err = errors.Wrap(err)
		return co, item, err
	}

	if alreadyCheckedOut && !store.shouldCheckOut(options, co, true) {
		if err = store.WriteFSItemToExternal(item, co.GetSkuExternal()); err != nil {
			err = errors.Wrap(err)
			return co, item, err
		}

		// FSItem does not have the object ID for certain so we need to add it to the
		// external on checkout
		co.GetSkuExternal().GetObjectId().ResetWith(co.GetSku().GetObjectId())
		co.SetState(checked_out_state.CheckedOut)

		return co, item, err
	}

	// ui.DebugBatsTestBody().Print(sku_fmt_debug.String(co.GetSku()))

	if err = store.checkoutOneForReal(
		options,
		co,
		item,
	); err != nil {
		err = errors.Wrap(err)
		return co, item, err
	}

	// FSItem does not have the object ID for certain so we need to add it to the
	// external on checkout
	co.GetSkuExternal().GetObjectId().ResetWith(co.GetSku().GetObjectId())

	return co, item, err
}

func (store *Store) prepareFSItemForCheckOut(
	options checkout_options.Options,
	co *sku.CheckedOut,
) (item *sku.FSItem, alreadyCheckedOut bool, err error) {
	fsOptions := GetCheckoutOptionsFromOptions(options)

	if store.config.IsDryRun() ||
		fsOptions.Path == PathOptionTempLocal {
		item = &sku.FSItem{}
		item.Reset()
		return item, alreadyCheckedOut, err
	}

	if item, alreadyCheckedOut = store.Get(co.GetSku().GetObjectId()); alreadyCheckedOut {
		if err = store.HydrateExternalFromItem(
			sku.CommitOptions{
				StoreOptions: sku.GetStoreOptionsRealizeSansProto(),
			},
			item,
			co.GetSku(),
			co.GetSkuExternal(),
		); err != nil {
			if sku.IsErrMergeConflict(err) && options.AllowConflicted {
				err = nil
			} else {
				err = errors.Wrap(err)
				return item, alreadyCheckedOut, err
			}
		}
	} else {
		if item, err = store.ReadFSItemFromExternal(co.GetSkuExternal()); err != nil {
			err = errors.Wrap(err)
			return item, alreadyCheckedOut, err
		}
	}

	// sku.DetermineState(co, true)

	return item, alreadyCheckedOut, err
}
