package env_workspace

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/mike/store_workspace"
)

// import "code.linenisgreat.com/dodder/go/src/juliett/sku"

// type ExternalStore interface {
// 	sku.ExternalStoreReadAllExternalItems
// 	sku.ExternalStoreUpdateTransacted
// 	sku.ExternalStoreReadExternalLikeFromObjectIdLike
// 	QueryCheckedOut
// }

type Store struct {
	store_workspace.Supplies
	store_workspace.StoreLike

	didInit   bool
	onceInit  sync.Once
	initError error
}

func (ve *Store) Initialize() (err error) {
	ve.onceInit.Do(func() {
		ve.initError = ve.StoreLike.Initialize(ve.Supplies)
		ve.didInit = true
	})

	err = ve.initError

	return
}

func (ve *Store) Flush() (err error) {
	if !ve.didInit {
		return
	}

	if err = ve.StoreLike.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) QueryCheckedOut(
	qg *query.Query,
	f interfaces.FuncIter[sku.SkuType],
) (err error) {
	es, ok := s.StoreLike.(store_workspace.QueryCheckedOut)

	if !ok {
		err = makeErrUnsupportedOperation(s, &s)
		return
	}

	if err = s.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = es.QueryCheckedOut(
		qg,
		f,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (es *Store) ReadAllExternalItems() (err error) {
	esado, ok := es.StoreLike.(interfaces.WorkspaceStoreReadAllExternalItems)

	if !ok {
		err = errors.ErrorWithStackf("store does not support %T", &esado)
		return
	}

	if err = es.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = esado.ReadAllExternalItems(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) ReadTransactedFromObjectId(
	o sku.CommitOptions,
	k1 interfaces.ObjectId,
	t *sku.Transacted,
) (e sku.ExternalLike, err error) {
	es, ok := s.StoreLike.(sku.ExternalStoreReadExternalLikeFromObjectIdLike)

	if !ok {
		err = makeErrUnsupportedOperation(s, &s)
		return
	}

	if err = s.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if e, err = es.ReadExternalLikeFromObjectIdLike(
		o,
		k1,
		t,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) ReadExternalLikeFromObjectIdLike(
	o sku.CommitOptions,
	k1 interfaces.Stringer,
	t *sku.Transacted,
) (e sku.ExternalLike, err error) {
	es, ok := s.StoreLike.(sku.ExternalStoreReadExternalLikeFromObjectIdLike)

	if !ok {
		err = makeErrUnsupportedOperation(s, &s)
		return
	}

	if err = s.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if e, err = es.ReadExternalLikeFromObjectIdLike(
		o,
		k1,
		t,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) CheckoutOne(
	options checkout_options.Options,
	sz sku.TransactedGetter,
) (cz sku.SkuType, err error) {
	es, ok := s.StoreLike.(store_workspace.CheckoutOne)

	if !ok {
		err = makeErrUnsupportedOperation(s, &s)
		return
	}

	if err = s.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if cz, err = es.CheckoutOne(
		options,
		sz,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) DeleteCheckedOut(el *sku.CheckedOut) (err error) {
	es, ok := s.StoreLike.(store_workspace.DeleteCheckedOut)

	if !ok {
		err = makeErrUnsupportedOperation(s, &s)
		return
	}

	if err = s.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = es.DeleteCheckedOut(el); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// Takes a given sku.Transacted (internal) and updates it with the state of its
// checked out counterpart (external).
func (s *Store) UpdateTransacted(z *sku.Transacted) (err error) {
	es, ok := s.StoreLike.(store_workspace.UpdateTransacted)

	if !ok {
		err = makeErrUnsupportedOperation(s, &s)
		return
	}

	if err = s.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = es.UpdateTransacted(z); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) UpdateTransactedFromBlobs(z sku.ExternalLike) (err error) {
	es, ok := s.StoreLike.(store_workspace.UpdateTransactedFromBlobs)

	if !ok {
		err = makeErrUnsupportedOperation(s, &es)
		return
	}

	if err = s.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = es.UpdateTransactedFromBlobs(z); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (es *Store) GetObjectIdsForString(
	v string,
) (k []sku.ExternalObjectId, err error) {
	if es == nil {
		err = collections.MakeErrNotFoundString(v)
		return
	}

	if err = es.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if k, err = es.StoreLike.GetObjectIdsForString(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) Open(
	m checkout_mode.Mode,
	ph interfaces.FuncIter[string],
	zsc sku.SkuTypeSet,
) (err error) {
	es, ok := s.StoreLike.(store_workspace.Open)

	if !ok {
		err = makeErrUnsupportedOperation(s, &es)
		return
	}

	if err = s.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = es.Open(m, ph, zsc); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) SaveBlob(el sku.ExternalLike) (err error) {
	es, ok := s.StoreLike.(sku.BlobSaver)

	if !ok {
		err = makeErrUnsupportedOperation(s, &es)
		return
	}

	if err = s.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = es.SaveBlob(el); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) UpdateCheckoutFromCheckedOut(
	options checkout_options.OptionsWithoutMode,
	col sku.SkuType,
) (err error) {
	es, ok := s.StoreLike.(store_workspace.UpdateCheckoutFromCheckedOut)

	if !ok {
		err = makeErrUnsupportedOperation(s, &s)
		return
	}

	if err = s.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = es.UpdateCheckoutFromCheckedOut(options, col); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) ReadCheckedOutFromTransacted(
	sk *sku.Transacted,
) (co *sku.CheckedOut, err error) {
	es, ok := s.StoreLike.(store_workspace.ReadCheckedOutFromTransacted)

	if !ok {
		err = makeErrUnsupportedOperation(s, &s)
		return
	}

	if err = s.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if co, err = es.ReadCheckedOutFromTransacted(sk); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) Merge(
	conflicted sku.Conflicted,
) (err error) {
	storeLike, ok := store.StoreLike.(store_workspace.Merge)

	if !ok {
		err = makeErrUnsupportedOperation(store, &store)
		return
	}

	if err = store.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = storeLike.Merge(
		conflicted,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) MergeCheckedOut(
	co *sku.CheckedOut,
	parentNegotiator sku.ParentNegotiator,
	allowMergeConflicts bool,
) (commitOptions sku.CommitOptions, err error) {
	es, ok := s.StoreLike.(store_workspace.MergeCheckedOut)

	if !ok {
		err = makeErrUnsupportedOperation(s, &s)
		return
	}

	if err = s.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if commitOptions, err = es.MergeCheckedOut(
		co,
		parentNegotiator,
		allowMergeConflicts,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
