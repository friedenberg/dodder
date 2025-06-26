package store_fs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (s *Store) checkoutOneIfNecessary(
	options checkout_options.Options,
	tg sku.TransactedGetter,
) (co *sku.CheckedOut, item *sku.FSItem, err error) {
	internal := tg.GetSku()
	co = GetCheckedOutPool().Get()

	sku.Resetter.ResetWith(co.GetSku(), internal)

	var alreadyCheckedOut bool

	if item, alreadyCheckedOut, err = s.prepareFSItemForCheckOut(options, co); err != nil {
		err = errors.Wrap(err)
		return
	}

	if alreadyCheckedOut && !s.shouldCheckOut(options, co, true) {
		if err = s.WriteFSItemToExternal(item, co.GetSkuExternal()); err != nil {
			err = errors.Wrap(err)
			return
		}

		// FSItem does not have the object ID for certain so we need to add it to the
		// external on checkout
		co.GetSkuExternal().GetObjectId().ResetWith(co.GetSku().GetObjectId())
		co.SetState(checked_out_state.CheckedOut)

		return
	}

	// ui.DebugBatsTestBody().Print(sku_fmt_debug.String(co.GetSku()))

	if err = s.checkoutOneForReal(
		options,
		co,
		item,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	// FSItem does not have the object ID for certain so we need to add it to the
	// external on checkout
	co.GetSkuExternal().GetObjectId().ResetWith(co.GetSku().GetObjectId())

	return
}

func (s *Store) prepareFSItemForCheckOut(
	options checkout_options.Options,
	co *sku.CheckedOut,
) (item *sku.FSItem, alreadyCheckedOut bool, err error) {
	fsOptions := GetCheckoutOptionsFromOptions(options)

	if s.config.IsDryRun() ||
		fsOptions.Path == PathOptionTempLocal {
		item = &sku.FSItem{}
		item.Reset()
		return
	}

	if item, alreadyCheckedOut = s.Get(co.GetSku().GetObjectId()); alreadyCheckedOut {
		if err = s.HydrateExternalFromItem(
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
				return
			}
		}
	} else {
		if item, err = s.ReadFSItemFromExternal(co.GetSkuExternal()); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	// sku.DetermineState(co, true)

	return
}
