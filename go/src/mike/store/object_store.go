package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (s *Store) GetObjectStore() sku.RepoStore {
	return s
}

func (s *Store) ReadPrimitiveQuery(
	qg sku.PrimitiveQueryGroup,
	w interfaces.FuncIter[*sku.Transacted],
) (err error) {
	return s.GetStreamIndex().ReadPrimitiveQuery(qg, w)
}
