package store

import (
	"slices"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	pkg_query "code.linenisgreat.com/dodder/go/src/kilo/query"
)

// TODO make iterator
func (store *Store) QueryPrimitive(
	qg sku.PrimitiveQueryGroup,
	f interfaces.FuncIter[*sku.Transacted],
) (err error) {
	e := pkg_query.MakeExecutorPrimitive(
		qg,
		store.GetStreamIndex().ReadPrimitiveQuery,
		store.ReadOneInto,
	)

	if err = e.ExecuteTransacted(f); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO make iterator
func (store *Store) QueryTransacted(
	qg *pkg_query.Query,
	output interfaces.FuncIter[*sku.Transacted],
) (err error) {
	var e pkg_query.Executor

	if e, err = store.makeQueryExecutor(qg); err != nil {
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

// TODO make iterator
func (store *Store) QueryTransactedAsSkuType(
	qg *pkg_query.Query,
	f interfaces.FuncIter[sku.SkuType],
) (err error) {
	var e pkg_query.Executor

	if e, err = store.makeQueryExecutor(qg); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = e.ExecuteTransactedAsSkuType(f); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO make iterator
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

func (store *Store) MakeBlobDigestBytesMap() (blobShaBytes map[string][]string, err error) {
	blobShaBytes = make(map[string][]string)
	var l sync.Mutex

	if err = store.QueryPrimitive(
		sku.MakePrimitiveQueryGroup(),
		func(sk *sku.Transacted) (err error) {
			l.Lock()
			defer l.Unlock()

			digestBytes := sk.Metadata.GetBlobDigest().GetBytes()
			oids := blobShaBytes[string(digestBytes)]
			oid := sk.ObjectId.String()
			loc, found := slices.BinarySearch(oids, oid)

			if found {
				return
			}

			oids = slices.Insert(oids, loc, oid)

			blobShaBytes[string(sk.Metadata.GetBlobDigest().GetBytes())] = oids

			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
