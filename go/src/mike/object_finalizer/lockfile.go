package object_finalizer

import (
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/object_metadata"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

func (finalizer finalizer) writeTypeLockIfNecessary(
	metadata object_metadata.IMetadataMutable,
	tipe ids.Type,
	funcs ...sku.FuncReadOne,
) (err errTypeLockfile) {
	// TODO stop excluding builtin types and create a process for signing those
	// too
	if tipe.IsEmpty() {
		err = ErrEmptyType
		return err
	} else if ids.IsBuiltin(tipe) {
		err = ErrBuiltinType
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

	if ok := sku.ReadOneObjectIdBespoke(tipe, typeObject, funcs...); ok {
		typeLock.ResetWithMarklId(typeObject.GetMetadataMutable().GetObjectSig())
	} else {
		err = ErrFailedToReadCurrentTypeObject
		return err
	}

	return err
}
