package object_finalizer

import (
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/object_metadata"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

func (finalizer finalizer) writeTypeLockIfNecessary(
	metadata object_metadata.IMetadataMutable,
	tipe ids.Type,
	funcs ...sku.FuncReadOne,
) (err error) {
	if tipe.IsEmpty() {
		err = ErrEmptyType
		return err
	} else if ids.IsBuiltin(tipe) {
		// TODO stop excluding builtin types and create a process for signing those
		// too
		err = ErrBuiltinType
		return err
	}

	typeLock := metadata.GetTypeLockMutable()

	// TODO There are cases where we will want to overwrite the typelock id,
	// should we use CommitOptions?
	if !typeLock.Value.IsNull() {
		return err
	}

	typeObject, repool := sku.GetTransactedPool().GetWithRepool()
	defer repool()

	if ok := sku.ReadOneObjectIdBespoke(tipe, typeObject, funcs...); ok {
		typeLock.Value.ResetWithMarklId(typeObject.GetMetadataMutable().GetObjectSig())
	} else {
		err = ErrFailedToReadCurrentTypeObject
		return err
	}

	return err
}
