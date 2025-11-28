package store_browser

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

func (store *Store) DeleteCheckedOut(co *sku.CheckedOut) (err error) {
	external := co.GetSkuExternal()

	var item Item

	if err = item.ReadFromExternal(external); err != nil {
		err = errors.Wrap(err)
		return err
	}

	item.ExternalId = external.GetSkuExternal().GetExternalObjectId().String()

	store.deleted[item.Url.Url()] = append(store.deleted[item.Url.Url()], checkedOutWithItem{
		CheckedOut: co.Clone(),
		Item:       item,
	})

	return err
}
