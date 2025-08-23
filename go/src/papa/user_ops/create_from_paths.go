package user_ops

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/script_value"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

type CreateFromPaths struct {
	*local_working_copy.Repo
	Proto      sku.Proto
	TextParser object_metadata.TextParser
	Filter     script_value.ScriptValue
	Delete     bool
	// ReadHinweisFromPath bool
}

func (op CreateFromPaths) Run(
	args ...string,
) (results sku.TransactedMutableSet, err error) {
	toCreate := make(map[string]*sku.Transacted)
	toDelete := fd.MakeMutableSet()

	commitOptions := sku.CommitOptions{
		StoreOptions: sku.GetStoreOptionsRealizeWithProto(),
	}

	for _, arg := range args {
		var object *sku.Transacted
		var fsItem sku.FSItem

		fsItem.Reset()

		fsItem.ExternalObjectId.SetGenre(genres.Zettel)

		if err = fsItem.Object.Set(arg); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = fsItem.FDs.Add(&fsItem.Object); err != nil {
			err = errors.Wrap(err)
			return
		}

		if object, err = op.GetEnvWorkspace().GetStoreFS().ReadExternalFromItem(
			commitOptions,
			&fsItem,
			nil,
		); err != nil {
			err = errors.ErrorWithStackf(
				"zettel text format error for path: %s: %s",
				arg,
				err,
			)
			return
		}

		digestWithoutTai := &object.Metadata.SelfWithoutTai

		if digestWithoutTai.IsNull() {
			return
		}

		digestBytes := digestWithoutTai.GetBytes()
		existing, ok := toCreate[string(digestBytes)]

		if ok {
			if err = existing.Metadata.Description.Set(
				object.Metadata.Description.String(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		} else {
			toCreate[string(digestBytes)] = object
		}

		if op.Delete {
			{
				var fdObject *fd.FD

				if fdObject, err = op.GetEnvWorkspace().GetStoreFS().GetObjectOrError(object); err != nil {
					err = errors.Wrap(err)
					return
				}

				toDelete.Add(fdObject)
			}

			{
				var fdBlob *fd.FD

				if fdBlob, err = op.GetEnvWorkspace().GetStoreFS().GetObjectOrError(object); err != nil {
					err = errors.Wrap(err)
					return
				}

				toDelete.Add(fdBlob)
			}
		}
	}

	results = sku.MakeTransactedMutableSet()

	if err = op.Lock(); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, object := range toCreate {
		if object.Metadata.IsEmpty() {
			return
		}

		op.Proto.Apply(object, genres.Zettel)

		if err = op.GetStore().CreateOrUpdateDefaultProto(
			object,
			sku.StoreOptions{
				ApplyProto: true,
			},
		); err != nil {
			// TODO-P2 add file for error handling
			op.handleStoreError(object, "", err)
			err = nil
			continue
		}

		results.Add(object)
	}

	for fdToDelete := range toDelete.All() {
		// TODO-P2 move to checkout store
		if err = op.GetEnvRepo().Delete(fdToDelete.GetPath()); err != nil {
			err = errors.Wrap(err)
			return
		}

		pathRel := op.GetEnvRepo().RelToCwdOrSame(fdToDelete.GetPath())

		// TODO-P2 move to printer
		op.GetUI().Printf("[%s] (deleted)", pathRel)
	}

	if err = op.Unlock(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (op CreateFromPaths) handleStoreError(
	object *sku.Transacted,
	path string,
	in error,
) {
	var err error

	var normalError interfaces.ErrorStackTracer

	if errors.As(in, &normalError) {
		ui.Err().Printf("%s", normalError.Error())
	} else {
		err = errors.ErrorWithStackf("writing zettel failed: %s: %s", path, in)
		ui.Err().Print(err)
	}
}
