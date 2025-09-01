package inventory_list_store

import (
	"sort"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/file_lock"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
	"code.linenisgreat.com/dodder/go/src/kilo/inventory_list_coders"
)

var ErrEmptyInventoryList = errors.New("empty inventory list")

type Store struct {
	lock sync.Mutex

	envRepo      env_repo.Env
	lockSmith    interfaces.LockSmith
	storeVersion interfaces.StoreVersion
	clock        ids.Clock

	inventoryListBlobStore
	blobBlobStore interfaces.BlobStore

	box *box_format.BoxTransacted

	ui sku.UIStorePrinters
}

var _ sku.InventoryListStore = &Store{}

type inventoryListBlobStore interface {
	interfaces.BlobStore

	getType() ids.Type
	getFormat() sku.ListCoder
	GetInventoryListCoderCloset() inventory_list_coders.Closet

	ReadOneBlobId(interfaces.MarklId) (*sku.Transacted, error)
	WriteInventoryListObject(*sku.Transacted) error

	AllInventoryLists() interfaces.SeqError[*sku.Transacted]
}

func (store *Store) Initialize(
	envRepo env_repo.Env,
	clock ids.Clock,
	inventoryListCoderCloset inventory_list_coders.Closet,
) (err error) {
	*store = Store{
		envRepo:       envRepo,
		lockSmith:     envRepo.GetLockSmith(),
		storeVersion:  envRepo.GetStoreVersion(),
		blobBlobStore: envRepo.GetDefaultBlobStore(),
		clock:         clock,
		box: box_format.MakeBoxTransactedArchive(
			envRepo,
			options_print.Options{}.WithPrintTai(true),
		),
	}

	blobType := ids.MustType(
		store.envRepo.GetConfigPublic().Blob.GetInventoryListTypeId(),
	)

	inventoryListBlobStore := envRepo.GetInventoryListBlobStore()
	coder := inventoryListCoderCloset.GetCoderForType(blobType)

	if store_version.LessOrEqual(
		store.storeVersion,
		store_version.V8,
	) {
		store.inventoryListBlobStore = &blobStoreV0{
			envRepo:                  envRepo,
			blobType:                 blobType,
			BlobStore:                inventoryListBlobStore,
			listFormat:               coder,
			inventoryListCoderCloset: inventoryListCoderCloset,
		}
	} else {
		store.inventoryListBlobStore = &blobStoreV1{
			envRepo:                  envRepo,
			pathLog:                  envRepo.FileInventoryListLog(),
			blobType:                 blobType,
			BlobStore:                inventoryListBlobStore,
			listFormat:               coder,
			inventoryListCoderCloset: inventoryListCoderCloset,
		}
	}

	return
}

func (store *Store) SetUIDelegate(ud sku.UIStorePrinters) {
	store.ui = ud
}

func (store *Store) GetEnv() env_ui.Env {
	return store.GetEnvRepo()
}

func (store *Store) Flush() (err error) {
	wg := errors.MakeWaitGroupParallel()
	return wg.GetError()
}

// TODO pass errors.Context
func (store *Store) FormatForVersion(
	storeVersion interfaces.StoreVersion,
) sku.ListCoder {
	tipe := ids.GetOrPanic(
		store.envRepo.GetConfigPublic().Blob.GetInventoryListTypeId(),
	).Type

	return store.GetInventoryListCoderCloset().GetCoderForType(tipe)
}

func (store *Store) GetTai() ids.Tai {
	if store.clock == nil {
		return ids.NowTai()
	} else {
		return store.clock.GetTai()
	}
}

func (store *Store) GetEnvRepo() env_repo.Env {
	return store.envRepo
}

func (store *Store) MakeOpenList() (openList *sku.OpenList, err error) {
	openList = &sku.OpenList{}

	if openList.Mover, err = store.blobBlobStore.Mover(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) AddObjectToOpenList(
	openList *sku.OpenList,
	object *sku.Transacted,
) (err error) {
	if err = object.FinalizeAndSignOverwrite(
		store.envRepo.GetConfigPrivate().Blob,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if _, err = inventory_list_coders.WriteObjectToOpenList(
		store.getFormat(),
		object,
		openList,
	); err != nil {
		err = errors.Wrapf(
			err,
			"%#v, format type: %q",
			object.Metadata.Fields,
			store.getType(),
		)

		return
	}

	return
}

func (store *Store) Create(
	openList *sku.OpenList,
) (object *sku.Transacted, err error) {
	if openList.Len == 0 {
		err = errors.Wrap(ErrEmptyInventoryList)
		return
	}

	if !store.lockSmith.IsAcquired() {
		err = file_lock.ErrLockRequired{
			Operation: "create inventory list",
		}

		return
	}

	object = sku.GetTransactedPool().Get()

	object.Metadata.Type = store.getType()
	object.Metadata.Description = openList.Description

	tai := store.GetTai()

	if err = object.ObjectId.SetWithIdLike(tai); err != nil {
		err = errors.Wrap(err)
		return
	}

	object.SetTai(tai)

	if err = openList.Mover.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	actual := openList.Mover.GetMarklId()
	// expected := &merkle.Id{}

	expected := object.GetBlobDigest()
	// if err = expected.SetMerkleId(merkle.HRPObjectBlobDigestSha256V0,
	// actual.GetBytes()); err != nil {
	// 	err = errors.Wrap(err)
	// 	return
	// }

	ui.Log().Print("expected", expected, "actual", actual)

	if expected.IsNull() {
		object.SetBlobDigest(actual)
	} else {
		if err = markl.MakeErrNotEqual(expected, actual); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	// if err = object.Metadata.GetRepoPubKeyMutable().SetMerkleId(
	// 	merkle.HRPRepoPubKeyV1,
	// 	store.envRepo.GetConfigPublic().Blob.GetPublicKey(),
	// ); err != nil {
	// 	err = errors.Wrap(err)
	// 	return
	// }

	if err = store.WriteInventoryListObject(object); err != nil {
		err = errors.Wrapf(err, "OpenList: %d", openList.Len)
		return
	}

	return
}

func (store *Store) WriteInventoryListBlob(
	remoteBlobStore interfaces.BlobStore,
	object *sku.Transacted,
	list *sku.ListTransacted,
) (err error) {
	if list.Len() == 0 {
		if !object.GetBlobDigest().IsNull() {
			err = errors.ErrorWithStackf(
				"inventory list has non-empty blob but passed in list is empty. %q",
				sku.String(object),
			)

			return
		}

		return
	}

	var writeCloser interfaces.WriteCloseMarklIdGetter

	if writeCloser, err = store.envRepo.GetDefaultBlobStore().BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, writeCloser)

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(writeCloser)
	defer repoolBufferedWriter()

	if _, err = store.GetInventoryListCoderCloset().WriteBlobToWriter(
		store.envRepo,
		object.GetType(),
		quiter.MakeSeqErrorFromSeq(list.All()),
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	actual := writeCloser.GetMarklId()
	expected := object.GetBlobDigest()

	ui.Log().Print("expected", expected, "actual", actual)

	if expected.IsNull() {
		object.SetBlobDigest(actual)
	} else {
		if err = markl.MakeErrNotEqual(expected, actual); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	// if !s.af.HasBlob(t.GetBlobSha()) {
	// 	err = errors.Errorf(
	// 		"inventory list blob missing after write (%d bytes, %d skus): %q",
	// 		n,
	// 		skus.Len(),
	// 		sku.String(t),
	// 	)

	// 	return
	// }

	// if _, _, err = s.blobStore.GetTransactedWithBlob(
	// 	t,
	// ); err != nil {
	// 	err = errors.Wrapf(err, "Blob Sha: %q", actual)
	// 	return
	// }

	return
}

func (store *Store) AllInventoryListContents(
	blobSha interfaces.MarklId,
) sku.Seq {
	return store.GetInventoryListCoderCloset().IterInventoryListBlobSkusFromBlobStore(
		store.getType(),
		store.blobBlobStore,
		blobSha,
	)
}

func (store *Store) ReadLast() (*sku.Transacted, error) {
	max := sku.GetTransactedPool().Get()

	for list, err := range store.AllInventoryLists() {
		if err != nil {
			return nil, errors.Wrap(err)
		}

		if sku.TransactedLessor.LessPtr(max, list) {
			sku.TransactedResetter.ResetWith(max, list)
		}
	}

	return max, nil
}

func (store *Store) AllInventoryListObjectsAndContents() interfaces.SeqError[sku.ObjectWithList] {
	return func(yield func(sku.ObjectWithList, error) bool) {
		var objectWithList sku.ObjectWithList

		for listObject, iterErr := range store.AllInventoryLists() {
			objectWithList.List = listObject
			objectWithList.Object = listObject

			if iterErr != nil {
				if !yield(objectWithList, iterErr) {
					return
				}
			}

			if !yield(objectWithList, nil) {
				return
			}

			iter := store.AllInventoryListContents(
				listObject.GetBlobDigest(),
			)

			for object, iterErr := range iter {
				objectWithList.Object = object

				if !yield(objectWithList, iterErr) {
					return
				}
			}
		}
	}
}

func (store *Store) AllInventoryListsSorted() sku.Seq {
	return func(yield func(*sku.Transacted, error) bool) {
		// TODO optimize space by storing digest and tai but not anything else
		var lists []*sku.Transacted

		for list, iterErr := range store.AllInventoryLists() {
			if iterErr != nil {
				if !yield(nil, errors.Wrap(iterErr)) {
					return
				}
			}

			lists = append(lists, list)
		}

		sort.Slice(
			lists,
			func(i, j int) bool { return lists[i].Less(lists[j]) },
		)

		for _, list := range lists {
			if !yield(list, nil) {
				return
			}
		}
	}
}
