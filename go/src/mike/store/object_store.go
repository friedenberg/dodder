package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (store *Store) GetObjectStore() sku.RepoStore {
	return store
}

func (store *Store) ReadPrimitiveQuery(
	qg sku.PrimitiveQueryGroup,
	w interfaces.FuncIter[*sku.Transacted],
) (err error) {
	return store.GetStreamIndex().ReadPrimitiveQuery(qg, w)
}
