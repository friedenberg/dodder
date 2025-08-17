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

	o := sku.CommitOptions{
		StoreOptions: sku.GetStoreOptionsRealizeWithProto(),
	}

	for _, arg := range args {
		var z *sku.Transacted
		var i sku.FSItem

		i.Reset()

		i.ExternalObjectId.SetGenre(genres.Zettel)

		if err = i.Object.Set(arg); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = i.FDs.Add(&i.Object); err != nil {
			err = errors.Wrap(err)
			return
		}

		if z, err = op.GetEnvWorkspace().GetStoreFS().ReadExternalFromItem(
			o,
			&i,
			nil,
		); err != nil {
			err = errors.ErrorWithStackf(
				"zettel text format error for path: %s: %s",
				arg,
				err,
			)
			return
		}

		sh := &z.Metadata.Digests.SelfMetadataWithoutTai

		if sh.IsNull() {
			return
		}

		digestBytes := sh.GetBytes()
		existing, ok := toCreate[string(digestBytes)]

		if ok {
			if err = existing.Metadata.Description.Set(
				z.Metadata.Description.String(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		} else {
			toCreate[string(digestBytes)] = z
		}

		if op.Delete {
			{
				var object *fd.FD

				if object, err = op.GetEnvWorkspace().GetStoreFS().GetObjectOrError(z); err != nil {
					err = errors.Wrap(err)
					return
				}

				var f fd.FD
				f.ResetWith(object)
				toDelete.Add(&f)
			}

			{
				var blob *fd.FD

				if blob, err = op.GetEnvWorkspace().GetStoreFS().GetObjectOrError(z); err != nil {
					err = errors.Wrap(err)
					return
				}

				var f fd.FD
				f.ResetWith(blob)
				toDelete.Add(&f)
			}
		}
	}

	results = sku.MakeTransactedMutableSet()

	if err = op.Lock(); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, z := range toCreate {
		if z.Metadata.IsEmpty() {
			return
		}

		op.Proto.Apply(z, genres.Zettel)

		if err = op.GetStore().CreateOrUpdateDefaultProto(
			z,
			sku.StoreOptions{
				ApplyProto: true,
			},
		); err != nil {
			// TODO-P2 add file for error handling
			op.handleStoreError(z, "", err)
			err = nil
			continue
		}

		results.Add(z)
	}

	for f := range toDelete.All() {
		// TODO-P2 move to checkout store
		if err = op.GetEnvRepo().Delete(f.GetPath()); err != nil {
			err = errors.Wrap(err)
			return
		}

		pathRel := op.GetEnvRepo().RelToCwdOrSame(f.GetPath())

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
