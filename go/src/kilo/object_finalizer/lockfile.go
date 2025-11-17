package object_finalizer

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/india/object_metadata"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

func (finalizer finalizer) writeTypeLockIfNecessary(
	metadata object_metadata.IMetadataMutable,
	tipe ids.Type,
	index sku.IndexPrimitives,
) (err error) {
	return
	// TODO stop excluding builtin types and create a process for signing those
	// too
	if tipe.IsEmpty() || ids.IsBuiltin(tipe) {
		return
	}

	lockfile := metadata.GetLockfileMutable()
	typeLock := lockfile.GetTypeLockMutable()

	// TODO There are cases where we will want to overwrite the typelock id,
	// should we use CommitOptions?
	if !typeLock.Id.IsNull() {
		return
	}

	typeObject, repool := sku.GetTransactedPool().GetWithRepool()
	defer repool()

	if err = index.ReadOneObjectId(
		tipe,
		typeObject,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	typeLock.Key = tipe.String()
	typeLock.Id.ResetWithMarklId(
		typeObject.GetMetadataMutable().GetObjectSig(),
	)

	return
}
