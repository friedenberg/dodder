package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (store *Store) tryNewHook(
	kinder *sku.Transacted,
	o sku.CommitOptions,
) (err error) {
	if !o.RunHooks {
		return
	}

	var t *sku.Transacted

	if t, err = store.ReadOneObjectId(kinder.GetType()); err != nil {
		if collections.IsErrNotFound(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	var blob type_blobs.Blob

	if blob, _, err = store.GetTypedBlobStore().Type.ParseTypedBlob(
		t.GetType(),
		t.GetBlobId(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer store.GetTypedBlobStore().Type.PutTypedBlob(t.GetType(), blob)

	script := blob.GetStringLuaHooks()

	if script == "" {
		return
	}

	if err = store.tryHookWithName(
		kinder,
		nil,
		o,
		t,
		script,
		"on_new",
	); err != nil {
		err = errors.Wrapf(err, "Hook: %#v", script)
		return
	}

	return
}

func (store *Store) TryFormatHook(
	object *sku.Transacted,
) (err error) {
	var objectMother *sku.Transacted

	if objectMother, err = store.ReadOneObjectId(object.GetObjectId()); err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	var objectType *sku.Transacted

	if objectType, err = store.ReadOneObjectId(object.GetType()); err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	var blob type_blobs.Blob

	if blob, _, err = store.GetTypedBlobStore().Type.ParseTypedBlob(
		objectType.GetType(),
		objectType.GetBlobId(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer store.GetTypedBlobStore().Type.PutTypedBlob(
		objectType.GetType(),
		blob,
	)

	script := blob.GetStringLuaHooks()

	if script == "" {
		return
	}

	if err = store.tryHookWithName(
		object,
		objectMother,
		sku.CommitOptions{},
		objectType,
		script,
		"on_format",
	); err != nil {
		err = errors.Wrapf(err, "Hook: %#v", script)
		return
	}

	return
}

func (store *Store) tryPreCommitHooks(
	kinder *sku.Transacted,
	mutter *sku.Transacted,
	o sku.CommitOptions,
) (err error) {
	if !o.RunHooks {
		return
	}

	type hook struct {
		script      string
		description string
	}

	hooks := []hook{}

	var t *sku.Transacted

	if t, err = store.ReadOneObjectId(kinder.GetType()); err != nil {
		if collections.IsErrNotFound(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	var blob type_blobs.Blob

	if blob, _, err = store.GetTypedBlobStore().Type.ParseTypedBlob(
		t.GetType(),
		t.GetBlobId(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer store.GetTypedBlobStore().Type.PutTypedBlob(t.GetType(), blob)

	script := blob.GetStringLuaHooks()

	hooks = append(hooks, hook{script: script, description: "type"})
	hooks = append(
		hooks,
		hook{
			script:      store.GetConfigStore().GetConfig().Hooks,
			description: "config-mutable",
		},
	)

	for _, h := range hooks {
		if h.script == "" {
			continue
		}

		if err = store.tryHookWithName(
			kinder,
			mutter,
			o,
			t,
			h.script,
			"on_pre_commit",
		); err != nil {
			err = errors.Wrapf(err, "Hook: %#v", h)
			err = errors.Wrapf(err, "Type: %q", kinder.GetType())

			if store.envRepo.Retry(
				"hook failed",
				"ignore error and continue?",
				err,
			) {
				// TODO fix this to properly continue past the failure
				err = nil
			} else {
				return
			}
		}
	}

	return
}

func (store *Store) tryPreCommitHook(
	kinder *sku.Transacted,
	mutter *sku.Transacted,
	storeOptions sku.StoreOptions,
	selbst *sku.Transacted,
	script string,
) (err error) {
	if !storeOptions.RunHooks || !storeOptions.AddToInventoryList {
		return
	}

	var vp sku.LuaVMPoolV1

	if vp, err = store.MakeLuaVMPoolV1(selbst, script); err != nil {
		err = errors.Wrap(err)
		return
	}

	var vm *sku.LuaVMV1

	if vm, err = vp.Get(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer vp.Put(vm)

	var tt *lua.LTable

	if tt, err = vm.GetTopTableOrError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	f := vm.GetField(tt, "on_pre_commit")

	if f.Type() != lua.LTFunction {
		return
	}

	tableKinder := vm.TablePool.Get()
	defer vm.TablePool.Put(tableKinder)

	sku.ToLuaTableV1(
		kinder,
		vm.LState,
		tableKinder,
	)

	var tableMutter *sku.LuaTableV1

	if mutter != nil {
		tableMutter = vm.TablePool.Get()
		defer vm.TablePool.Put(tableMutter)

		sku.ToLuaTableV1(
			mutter,
			vm.LState,
			tableMutter,
		)
	}

	vm.Push(f)
	vm.Push(tableKinder.Transacted)
	vm.Push(tableMutter.Transacted)

	if err = vm.PCall(2, 1, nil); err != nil {
		err = errors.Wrap(err)
		return
	}

	retval := vm.LState.Get(1)
	vm.Pop(1)

	if retval.Type() != lua.LTNil {
		err = errors.ErrorWithStackf("lua error: %s", retval)
		return
	}

	if err = sku.FromLuaTableV1(kinder, vm.LState, tableKinder); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO add method with hook with reader
func (store *Store) tryHookWithName(
	kinder *sku.Transacted,
	mutter *sku.Transacted,
	o sku.CommitOptions,
	self *sku.Transacted,
	script string,
	name string,
) (err error) {
	var vp sku.LuaVMPoolV1

	if vp, err = store.MakeLuaVMPoolV1(self, script); err != nil {
		err = errors.Wrap(err)
		return
	}

	var vm *sku.LuaVMV1

	if vm, err = vp.Get(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err != nil {
		return
	}

	defer vp.Put(vm)

	var tt *lua.LTable

	if tt, err = vm.GetTopTableOrError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	f := vm.GetField(tt, name)

	if f.Type() != lua.LTFunction {
		return
	}

	tableKinder := vm.TablePool.Get()
	defer vm.TablePool.Put(tableKinder)

	sku.ToLuaTableV1(
		kinder,
		vm.LState,
		tableKinder,
	)

	var tableMutter *sku.LuaTableV1

	if mutter != nil {
		tableMutter = vm.TablePool.Get()
		defer vm.TablePool.Put(tableMutter)

		sku.ToLuaTableV1(
			mutter,
			vm.LState,
			tableMutter,
		)
	}

	vm.Push(f)
	vm.Push(tableKinder.Transacted)

	if tableMutter != nil {
		vm.Push(tableMutter.Transacted)
	} else {
		vm.Push(nil)
	}

	if err = vm.PCall(2, 1, nil); err != nil {
		err = errors.Wrap(err)
		return
	}

	retval := vm.LState.Get(1)
	vm.Pop(1)

	if retval.Type() != lua.LTNil {
		err = errors.ErrorWithStackf("lua error: %s", retval)
		return
	}

	if err = sku.FromLuaTableV1(kinder, vm.LState, tableKinder); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
