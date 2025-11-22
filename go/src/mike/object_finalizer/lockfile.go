package object_finalizer

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/object_metadata"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

func (finalizer finalizer) writeTypeLockIfNecessary(
	metadata object_metadata.IMetadataMutable,
	tipe ids.Type,
	funcReadOne sku.FuncReadOne,
) (err error) {
	return err
	// TODO stop excluding builtin types and create a process for signing those
	// too
	if tipe.IsEmpty() || ids.IsBuiltin(tipe) {
		return err
	}

	lockfile := metadata.GetLockfileMutable()
	typeLock := lockfile.GetTypeMutable()

	// TODO There are cases where we will want to overwrite the typelock id,
	// should we use CommitOptions?
	if !typeLock.IsNull() {
		return err
	}

	typeObject, repool := sku.GetTransactedPool().GetWithRepool()
	defer repool()

	if !sku.ReadOneObjectIdBespoke(funcReadOne, tipe, typeObject) {
		panic(errors.Errorf("failed to read type"))
	}

	typeLock.ResetWithMarklId(typeObject.GetMetadataMutable().GetObjectSig())

	return err
}
