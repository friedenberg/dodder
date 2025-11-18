package store_browser

import (
	"net/url"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

// TODO decide how this should behave
func (store *Store) UpdateTransacted(sk *sku.Transacted) (err error) {
	if !sk.GetType().Equals(store.typ) {
		return err
	}

	var uSku *url.URL

	if uSku, err = store.getUrl(sk); err != nil {
		err = errors.Wrap(err)
		return err
	}

	_, ok := store.urls[*uSku]

	if !ok {
		return err
	}

	return err
}
