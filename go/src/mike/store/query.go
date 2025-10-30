package store

import (
	"slices"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	pkg_query "code.linenisgreat.com/dodder/go/src/kilo/queries"
)

// TODO make iterator
func (store *Store) QueryPrimitive(
	group sku.PrimitiveQueryGroup,
	funcIter interfaces.FuncIter[*sku.Transacted],
) (err error) {
	executor := pkg_query.MakeExecutorPrimitive(
		group,
		store.GetStreamIndex().ReadPrimitiveQuery,
		store.ReadOneInto,
	)

	if err = executor.ExecuteTransacted(funcIter); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO make iterator
func (store *Store) QueryTransacted(
	group *pkg_query.Query,
	funcIter interfaces.FuncIter[*sku.Transacted],
) (err error) {
	var executor pkg_query.Executor

	if executor, err = store.makeQueryExecutor(group); err != nil {
		err = errors.Wrap(err)
		return err
	}

	var object *sku.Transacted

	switch {
	case true:
		// TODO why does this not work with trying to read internal
		if object, err = executor.ExecuteExactlyOneExternalObject(false); err != nil {
			err = nil
			break
		}

		if err = funcIter(object); err != nil {
			err = errors.Wrap(err)
			return err
		}

		return err
	}

	if err = executor.ExecuteTransacted(funcIter); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO make iterator
func (store *Store) QueryTransactedAsSkuType(
	query *pkg_query.Query,
	funcIter interfaces.FuncIter[sku.SkuType],
) (err error) {
	var executor pkg_query.Executor

	if executor, err = store.makeQueryExecutor(query); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = executor.ExecuteTransactedAsSkuType(funcIter); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO make iterator
func (store *Store) QuerySkuType(
	query *pkg_query.Query,
	output interfaces.FuncIter[sku.SkuType],
) (err error) {
	var executor pkg_query.Executor

	if executor, err = store.makeQueryExecutor(query); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = executor.ExecuteSkuType(output); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (store *Store) QueryExactlyOneExternal(
	query *pkg_query.Query,
) (object *sku.Transacted, err error) {
	var executor pkg_query.Executor

	if executor, err = store.makeQueryExecutor(query); err != nil {
		err = errors.Wrap(err)
		return object, err
	}

	if object, err = executor.ExecuteExactlyOneExternalObject(true); err != nil {
		err = errors.Wrap(err)
		return object, err
	}

	return object, err
}

func (store *Store) QueryExactlyOne(
	queryGroup *pkg_query.Query,
) (object *sku.Transacted, err error) {
	var executor pkg_query.Executor

	if executor, err = store.makeQueryExecutor(queryGroup); err != nil {
		err = errors.Wrap(err)
		return object, err
	}

	if object, err = executor.ExecuteExactlyOne(); err != nil {
		err = errors.Wrap(err)
		return object, err
	}

	return object, err
}

func (store *Store) MakeBlobDigestObjectIdsMap() (blobDigestObjectIds map[string][]string, err error) {
	blobDigestObjectIds = make(map[string][]string)
	var lock sync.Mutex

	if err = store.QueryPrimitive(
		sku.MakePrimitiveQueryGroup(),
		func(object *sku.Transacted) (err error) {
			lock.Lock()
			defer lock.Unlock()

			digestBytes := object.Metadata.GetBlobDigest().GetBytes()
			objectIds := blobDigestObjectIds[string(digestBytes)]
			oid := object.ObjectId.String()
			loc, found := slices.BinarySearch(objectIds, oid)

			if found {
				return err
			}

			objectIds = slices.Insert(objectIds, loc, oid)

			bites := object.Metadata.GetBlobDigest().GetBytes()
			blobDigestObjectIds[string(bites)] = objectIds

			return err
		},
	); err != nil {
		err = errors.Wrap(err)
		return blobDigestObjectIds, err
	}

	return blobDigestObjectIds, err
}
