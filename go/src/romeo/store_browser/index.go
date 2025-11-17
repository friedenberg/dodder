package store_browser

import (
	"net/url"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

// TODO abstract and regenerate on commit / reindex
func (store *Store) initializeIndex() (err error) {
	if err = store.initializeCache(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	var l sync.Mutex

	if err = store.externalStoreInfo.ReadPrimitiveQuery(
		nil,
		func(sk *sku.Transacted) (err error) {
			if !sk.GetType().Equals(store.typ) {
				return err
			}

			var u *url.URL

			if u, err = store.getUrl(sk); err != nil {
				err = nil
				return err
			}

			cl := sku.GetTransactedPool().Get()
			sku.TransactedResetter.ResetWith(cl, sk)

			l.Lock()
			defer l.Unlock()

			{
				existing, ok := store.transactedUrlIndex[*u]

				if !ok {
					existing = sku.MakeTransactedMutableSet()
					store.transactedUrlIndex[*u] = existing
				}

				if err = existing.Add(cl); err != nil {
					err = errors.Wrap(err)
					return err
				}
			}

			{
				existing, ok := store.tabCache.Rows[sk.ObjectId.String()]

				if ok {
					store.transactedItemIndex[existing] = cl
				}
			}

			return err
		},
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
