package store

import (
	"slices"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	pkg_query "code.linenisgreat.com/dodder/go/src/kilo/query"
)

func (s *Store) QueryPrimitive(
	qg sku.PrimitiveQueryGroup,
	f interfaces.FuncIter[*sku.Transacted],
) (err error) {
	e := pkg_query.MakeExecutorPrimitive(
		qg,
		s.GetStreamIndex().ReadPrimitiveQuery,
		s.ReadOneInto,
	)

	if err = e.ExecuteTransacted(f); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) QueryTransacted(
	qg *pkg_query.Query,
	output interfaces.FuncIter[*sku.Transacted],
) (err error) {
	var e pkg_query.Executor

	if e, err = s.makeQueryExecutor(qg); err != nil {
		err = errors.Wrap(err)
		return
	}

	var sk *sku.Transacted

	switch {
	case true:
		// TODO why does this not work with trying to read internal
		if sk, err = e.ExecuteExactlyOneExternalObject(false); err != nil {
			err = nil
			break
		}

		if err = output(sk); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}

	if err = e.ExecuteTransacted(output); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) QueryTransactedAsSkuType(
	qg *pkg_query.Query,
	f interfaces.FuncIter[sku.SkuType],
) (err error) {
	var e pkg_query.Executor

	if e, err = s.makeQueryExecutor(qg); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = e.ExecuteTransactedAsSkuType(f); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) QuerySkuType(
	query *pkg_query.Query,
	output interfaces.FuncIter[sku.SkuType],
) (err error) {
	var executor pkg_query.Executor

	if executor, err = store.makeQueryExecutor(query); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = executor.ExecuteSkuType(output); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) QueryExactlyOneExternal(
	query *pkg_query.Query,
) (sk *sku.Transacted, err error) {
	var executor pkg_query.Executor

	if executor, err = store.makeQueryExecutor(query); err != nil {
		err = errors.Wrap(err)
		return
	}

	if sk, err = executor.ExecuteExactlyOneExternalObject(true); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) QueryExactlyOne(
	queryGroup *pkg_query.Query,
) (sk *sku.Transacted, err error) {
	var executor pkg_query.Executor

	if executor, err = store.makeQueryExecutor(queryGroup); err != nil {
		err = errors.Wrap(err)
		return
	}

	if sk, err = executor.ExecuteExactlyOne(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) MakeBlobShaBytesMap() (blobShaBytes map[sha.Bytes][]string, err error) {
	blobShaBytes = make(map[sha.Bytes][]string)
	var l sync.Mutex

	if err = s.QueryPrimitive(
		sku.MakePrimitiveQueryGroup(),
		func(sk *sku.Transacted) (err error) {
			l.Lock()
			defer l.Unlock()

			k := sk.Metadata.Blob.GetBytes()
			oids := blobShaBytes[k]
			oid := sk.ObjectId.String()
			loc, found := slices.BinarySearch(oids, oid)

			if found {
				return
			}

			oids = slices.Insert(oids, loc, oid)

			blobShaBytes[sk.Metadata.Blob.GetBytes()] = oids

			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
