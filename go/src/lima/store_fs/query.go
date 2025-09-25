package store_fs

import (
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
)

func (store *Store) QueryCheckedOut(
	queryGroup *query.Query,
	funk interfaces.FuncIter[sku.SkuType],
) (err error) {
	waitGroup := errors.MakeWaitGroupParallel()

	waitGroup.Do(func() (err error) {
		funcIterFSItems := store.makeFuncIterHydrateCheckedOutProbablyCheckedOut(
			store.makeFuncIterFilterAndApply(queryGroup, funk),
		)

		for item := range store.probablyCheckedOut.All() {
			if err = funcIterFSItems(item); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}

		return err
	})

	if !queryGroup.ExcludeUntracked {
		waitGroup.Do(func() (err error) {
			funcIterFSItems := store.makeFuncIterHydrateCheckedOutDefinitelyNotCheckedOut(
				store.makeFuncIterFilterAndApply(queryGroup, funk),
			)

			if err = store.queryUntracked(queryGroup, funcIterFSItems); err != nil {
				err = errors.Wrap(err)
				return err
			}

			return err
		})
	}

	if err = waitGroup.GetError(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (store *Store) makeFuncIterHydrateCheckedOutProbablyCheckedOut(
	out interfaces.FuncIter[sku.SkuType],
) interfaces.FuncIter[*sku.FSItem] {
	return func(item *sku.FSItem) (err error) {
		checkedOut := GetCheckedOutPool().Get()

		// at a bare minimum, the internal object ID must always be set as there
		// are hard assumptions about internal being valid throughout the
		// reading cycle
		if err = checkedOut.GetSku().ObjectId.SetObjectIdLike(
			&item.ExternalObjectId,
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

		hasInternal := true

		var objectId ids.ObjectId

		if err = objectId.SetObjectIdLike(item.GetExternalObjectId()); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = store.storeSupplies.ReadOneInto(
			&objectId,
			checkedOut.GetSku(),
		); err != nil {
			if collections.IsErrNotFound(err) ||
				genres.IsErrUnsupportedGenre(err) {
				hasInternal = false
				err = nil
			} else {
				err = errors.Wrap(err)
				return err
			}
		}

		if err = store.HydrateExternalFromItem(
			sku.CommitOptions{
				StoreOptions: sku.StoreOptions{
					UpdateTai: true,
				},
			},
			item,
			checkedOut.GetSku(),
			checkedOut.GetSkuExternal(),
		); err != nil {
			if sku.IsErrMergeConflict(err) {
				checkedOut.SetState(checked_out_state.Conflicted)

				if err = checkedOut.GetSkuExternal().ObjectId.SetWithIdLike(
					&checkedOut.GetSku().ObjectId,
				); err != nil {
					err = errors.Wrap(err)
					return err
				}
			} else {
				err = errors.Wrapf(err, "Cwd: %#v", item.Debug())
				return err
			}
		}

		if !item.Conflict.IsEmpty() {
			checkedOut.SetState(checked_out_state.Conflicted)
		} else if !hasInternal {
			checkedOut.SetState(checked_out_state.Untracked)
		} else {
			checkedOut.SetState(checked_out_state.CheckedOut)
		}

		if err = store.WriteFSItemToExternal(item, checkedOut.GetSkuExternal()); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = out(checkedOut); err != nil {
			err = errors.Wrap(err)
			return err
		}

		return err
	}
}

func (store *Store) makeFuncIterHydrateCheckedOutDefinitelyNotCheckedOut(
	f interfaces.FuncIter[sku.SkuType],
) interfaces.FuncIter[any] {
	return func(itemUnknown any) (err error) {
		co := sku.GetCheckedOutPool().Get()

		switch item := itemUnknown.(type) {
		case *sku.FSItem:
			if err = store.hydrateDefinitelyNotCheckedOutUnrecognizedItem(
				item,
				co,
				f,
			); err != nil {
				err = errors.Wrap(err)
				return err
			}

		case *fsItemRecognized:
			if err = store.hydrateDefinitelyNotCheckedOutRecognizedItem(
				item,
				co,
				f,
			); err != nil {
				err = errors.Wrap(err)
				return err
			}

		default:
			err = errors.ErrorWithStackf("unsupported type for item: %T", itemUnknown)
			return err
		}

		return err
	}
}

func (store *Store) hydrateDefinitelyNotCheckedOutUnrecognizedItem(
	item *sku.FSItem,
	co *sku.CheckedOut,
	f interfaces.FuncIter[sku.SkuType],
) (err error) {
	if !item.Conflict.IsEmpty() {
		err = errors.ErrorWithStackf(
			"cannot have a conflict for a definitely not checked out blob: %s",
			item.Debug(),
		)
		return err
	}

	if err = co.GetSku().ObjectId.SetObjectIdLike(
		&item.ExternalObjectId,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = co.GetSkuExternal().ObjectId.SetObjectIdLike(
		&item.ExternalObjectId,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = store.readOneExternalBlob(
		co.GetSkuExternal(),
		co.GetSku(),
		item,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = store.WriteFSItemToExternal(item, co.GetSkuExternal()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	co.SetState(checked_out_state.Untracked)

	if err = f(co); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (store *Store) hydrateDefinitelyNotCheckedOutRecognizedItem(
	item *fsItemRecognized,
	co *sku.CheckedOut,
	f interfaces.FuncIter[sku.SkuType],
) (err error) {
	sku.TransactedResetter.ResetWith(co.GetSku(), &item.Recognized)
	sku.TransactedResetter.ResetWith(co.GetSkuExternal(), &item.Recognized)

	co.SetState(checked_out_state.Recognized)

	for _, item := range item.Matching {
		if err = item.WriteToSku(
			co.GetSkuExternal(),
			store.envRepo,
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

		co.GetSkuExternal().ObjectId.SetGenre(genres.Blob)

		if err = store.WriteFSItemToExternal(item, co.GetSkuExternal()); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = f(co); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (store *Store) makeFuncIterFilterAndApply(
	qg *query.Query,
	f interfaces.FuncIter[sku.SkuType],
) interfaces.FuncIter[*sku.CheckedOut] {
	return func(co *sku.CheckedOut) (err error) {
		if !query.ContainsExternalSku(
			qg,
			co.GetSkuExternal(),
			co.GetState(),
		) {
			return err
		}

		if err = f(co); err != nil {
			err = errors.Wrap(err)
			return err
		}

		return err
	}
}

type fsItemRecognized struct {
	Recognized sku.Transacted
	Matching   []*sku.FSItem
}

func (store *Store) queryUntracked(
	qg *query.Query, // TODO use this to conditionally perform recognition
	aco interfaces.FuncIter[any],
) (err error) {
	definitelyNotCheckedOut := store.dirInfo.definitelyNotCheckedOut.Clone()

	// TODO move to initial parse?
	if err = definitelyNotCheckedOut.ConsolidateDuplicateBlobs(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	allRecognized := make([]*fsItemRecognized, 0)

	addRecognizedIfNecessary := func(
		object *sku.Transacted,
		digest interfaces.MutableMarklId,
		digestCache map[string]interfaces.MutableSetLike[*sku.FSItem],
	) (item *fsItemRecognized, err error) {
		if digest.IsNull() {
			return item, err
		}

		key := digest.GetBytes()
		recognized, ok := digestCache[string(key)]

		if !ok {
			return item, err
		}

		item = &fsItemRecognized{}

		sku.TransactedResetter.ResetWith(&item.Recognized, object)

		for recognized := range recognized.All() {
			item.Matching = append(item.Matching, recognized)
		}

		return item, err
	}

	if err = store.storeSupplies.ReadPrimitiveQuery(
		nil,
		func(object *sku.Transacted) (err error) {
			var recognizedBlob, recognizedObject *fsItemRecognized

			if recognizedBlob, err = addRecognizedIfNecessary(
				object,
				object.Metadata.GetBlobDigestMutable(),
				definitelyNotCheckedOut.digests,
			); err != nil {
				err = errors.Wrap(err)
				return err
			}

			if recognizedObject, err = addRecognizedIfNecessary(
				object,
				&object.Metadata.SelfWithoutTai,
				store.probablyCheckedOut.digests,
			); err != nil {
				err = errors.Wrap(err)
				return err
			}

			if recognizedBlob != nil {
				allRecognized = append(allRecognized, recognizedBlob)

				for _, item := range recognizedBlob.Matching {
					definitelyNotCheckedOut.Del(item)
				}
			}

			if recognizedObject != nil {
				allRecognized = append(allRecognized, recognizedObject)

				for _, item := range recognizedObject.Matching {
					definitelyNotCheckedOut.Del(item)
				}
			}

			return err
		},
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	{
		blobs := make([]*sku.FSItem, 0, definitelyNotCheckedOut.Len())

		for fds := range definitelyNotCheckedOut.All() {
			blobs = append(blobs, fds)
		}

		sort.Slice(
			blobs,
			func(i, j int) bool {
				return blobs[i].ExternalObjectId.String() < blobs[j].ExternalObjectId.String()
			},
		)

		for _, fds := range blobs {
			// if fds.State == external_state.Recognized {
			// 	continue
			// }

			if err = aco(fds); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}

	}

	// if false {
	// 	objects := make([]*sku.FSItem, 0, s.dirItems.probablyCheckedOut.Len())

	// 	if err = s.dirItems.probablyCheckedOut.Each(
	// 		func(fds *sku.FSItem) (err error) {
	// 			objects = append(objects, fds)
	// 			return
	// 		},
	// 	); err != nil {
	// 		err = errors.Wrap(err)
	// 		return
	// 	}

	// 	sort.Slice(
	// 		objects,
	// 		func(i, j int) bool {
	// 			return objects[i].ExternalObjectId.String() <
	// objects[j].ExternalObjectId.String()
	// 		},
	// 	)

	// 	for _, fds := range objects {
	// 		// if fds.State == external_state.Recognized {
	// 		// 	continue
	// 		// }

	// 		if err = aco(fds); err != nil {
	// 			err = errors.Wrap(err)
	// 			return
	// 		}
	// 	}
	// }

	for _, fds := range allRecognized {
		if err = aco(fds); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}
