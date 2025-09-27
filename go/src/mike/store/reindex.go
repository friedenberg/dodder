package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/file_lock"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/stream_index"
)

// TODO-P2 add support for quiet reindexing
func (store *Store) Reindex() (err error) {
	if !store.GetEnvRepo().GetLockSmith().IsAcquired() {
		err = file_lock.ErrLockRequired{
			Operation: "reindex",
		}

		return err
	}

	if err = store.ResetIndexes(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = store.GetEnvRepo().ResetCache(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	var reindexer stream_index.IndexCommon

	if reindexer, err = store.streamIndex.MakeReindexer(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	commitFacilitator := commitFacilitator{
		Store: store,
		index: reindexer,
	}

	type objectWithError struct {
		error
		sku.ObjectWithList
	}

	objectsWithErrors := make(map[string]objectWithError)
	unidentifiedErrors := make([]error, 0)

	seq := store.GetInventoryListStore().AllInventoryListObjectsAndContents()

	for objectWithList, iterErr := range seq {
		if iterErr != nil {
			if objectWithList.List == nil {
				unidentifiedErrors = append(unidentifiedErrors, iterErr)
			} else {
				keyBytes := objectWithList.List.GetObjectDigest().GetBytes()

				objectsWithErrors[string(keyBytes)] = objectWithError{
					error: iterErr,
					ObjectWithList: sku.ObjectWithList{
						List: objectWithList.List.CloneTransacted(),
					},
				}
			}

			continue
		}

		if objectWithList.Object == nil {
			panic("empty object")
		}

		if err = store.reindexOne(commitFacilitator, objectWithList); err != nil {
			keyBytes := objectWithList.List.GetObjectDigest().GetBytes()

			objectsWithErrors[string(keyBytes)] = objectWithError{
				error: err,
				ObjectWithList: sku.ObjectWithList{
					Object: objectWithList.Object.CloneTransacted(),
					List:   objectWithList.List.CloneTransacted(),
				},
			}

			continue
		}
	}

	store.envRepo.GetUI().Print("unidentified errors:")

	for _, err := range unidentifiedErrors {
		ui.CLIErrorTreeEncoder.EncodeTo(err, store.envRepo.GetUI())
	}

	store.envRepo.GetUI().Print("objects with errors:")

	for _, objectWithError := range objectsWithErrors {
		ui.CLIErrorTreeEncoder.EncodeTo(err, store.envRepo.GetUI())

		if objectWithError.Object == nil {
			store.envRepo.GetUI().Printf(
				"Error: %s, List: %q",
				objectWithError.error,
				sku.String(objectWithError.List),
			)
		} else {
			store.envRepo.GetUI().Printf(
				"Error: %s, List: %q, Object: %q",
				objectWithError.error,
				sku.String(objectWithError.List),
				sku.String(objectWithError.Object),
			)
		}
	}

	return err
}

func (store *Store) reindexOne(
	commitFacilitator commitFacilitator,
	object sku.ObjectWithList,
) (err error) {
	options := sku.CommitOptions{
		StoreOptions: sku.GetStoreOptionsReindex(),
	}

	if err = commitFacilitator.commit(object.Object, options); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = store.GetAbbrStore().AddObjectToIdIndex(
		object.Object,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
