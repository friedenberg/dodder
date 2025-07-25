package user_ops

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
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

func (c CreateFromPaths) Run(
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

		if err = i.MutableSetLike.Add(&i.Object); err != nil {
			err = errors.Wrap(err)
			return
		}

		if z, err = c.GetEnvWorkspace().GetStoreFS().ReadExternalFromItem(
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

		sh := &z.Metadata.Shas.SelfMetadataWithoutTai

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

		if c.Delete {
			{
				var object *fd.FD

				if object, err = c.GetEnvWorkspace().GetStoreFS().GetObjectOrError(z); err != nil {
					err = errors.Wrap(err)
					return
				}

				var f fd.FD
				f.ResetWith(object)
				toDelete.Add(&f)
			}

			{
				var blob *fd.FD

				if blob, err = c.GetEnvWorkspace().GetStoreFS().GetObjectOrError(z); err != nil {
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

	if err = c.Lock(); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, z := range toCreate {
		if z.Metadata.IsEmpty() {
			return
		}

		c.Proto.Apply(z, genres.Zettel)

		if err = c.GetStore().CreateOrUpdateDefaultProto(
			z,
			sku.StoreOptions{
				ApplyProto: true,
			},
		); err != nil {
			// TODO-P2 add file for error handling
			c.handleStoreError(z, "", err)
			err = nil
			continue
		}

		results.Add(z)
	}

	for f := range toDelete.All() {
		// TODO-P2 move to checkout store
		if err = c.GetEnvRepo().Delete(f.GetPath()); err != nil {
			err = errors.Wrap(err)
			return
		}

		pathRel := c.GetEnvRepo().RelToCwdOrSame(f.GetPath())

		// TODO-P2 move to printer
		c.GetUI().Printf("[%s] (deleted)", pathRel)
	}

	if err = c.Unlock(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (c CreateFromPaths) handleStoreError(
	z *sku.Transacted,
	f string,
	in error,
) {
	var err error

	var normalError errors.StackTracer

	if errors.As(in, &normalError) {
		ui.Err().Printf("%s", normalError.Error())
	} else {
		err = errors.ErrorWithStackf("writing zettel failed: %s: %s", f, in)
		ui.Err().Print(err)
	}
}
