package inventory_list_store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/comments"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (store *Store) Lock() error {
	return store.envRepo.GetLockSmith().Lock()
}

func (store *Store) Unlock() error {
	return store.envRepo.GetLockSmith().Unlock()
}

func (store *Store) Commit(
	externalLike sku.ExternalLike,
	_ sku.CommitOptions,
) (err error) {
	sk := externalLike.GetSku()

	if sk.GetGenre() != genres.InventoryList {
		err = genres.MakeErrUnsupportedGenre(sk.GetGenre())
		return
	}

	// TODO transform this inventory list into a local inventory list and update
	// its tai
	if err = store.WriteInventoryListObject(sk); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.ui.TransactedNew(sk); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) ReadOneInto(
	oid interfaces.ObjectId,
	_ *sku.Transacted,
) (err error) {
	if oid.GetGenre() != genres.InventoryList {
		err = genres.MakeErrUnsupportedGenre(oid.GetGenre())
		return
	}

	err = comments.Implement()
	// err = errors.BadRequestf("%q", oid)

	return
}

// TODO
func (store *Store) ReadPrimitiveQuery(
	queryGroup sku.PrimitiveQueryGroup,
	output interfaces.FuncIter[*sku.Transacted],
) (err error) {
	if err = store.ReadAllSkus(
		func(_, sk *sku.Transacted) (err error) {
			if err = output(sk); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
