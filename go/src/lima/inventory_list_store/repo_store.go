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
	return errors.Err405MethodNotAllowed
}

func (store *Store) ReadOneInto(
	objectId interfaces.ObjectId,
	_ *sku.Transacted,
) (err error) {
	if objectId.GetGenre() != genres.InventoryList {
		err = genres.MakeErrUnsupportedGenre(objectId.GetGenre())
		return
	}

	err = comments.Implement()

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
