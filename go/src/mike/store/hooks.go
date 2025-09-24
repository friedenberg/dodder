package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/sku_lua"
)

func (store *Store) tryNewHook(
	child *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	if !options.RunHooks {
		return err
	}

	var typeObject *sku.Transacted

	if typeObject, err = store.ReadOneObjectId(child.GetType()); err != nil {
		if collections.IsErrNotFound(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return err
	} else if typeObject == nil {
		return err
	}

	var blob type_blobs.Blob
	var repool interfaces.FuncRepool

	if blob, repool, _, err = store.GetTypedBlobStore().Type.ParseTypedBlob(
		typeObject.GetType(),
		typeObject.GetBlobDigest(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer repool()

	script := blob.GetStringLuaHooks()

	if script == "" {
		return err
	}

	if err = store.tryHookWithName(
		child,
		nil,
		options,
		typeObject,
		script,
		"on_new",
	); err != nil {
		err = errors.Wrapf(err, "Hook: %#v", script)
		return err
	}

	return err
}

func (store *Store) TryFormatHook(
	object *sku.Transacted,
) (err error) {
	var objectMother *sku.Transacted

	if objectMother, err = store.ReadOneObjectId(object.GetObjectId()); err != nil {
		if collections.IsErrNotFound(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return err
		}
	}

	var objectType *sku.Transacted

	if objectType, err = store.ReadOneObjectId(object.GetType()); err != nil {
		if collections.IsErrNotFound(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return err
		}
	}

	var blob type_blobs.Blob
	var repool interfaces.FuncRepool

	if blob, repool, _, err = store.GetTypedBlobStore().Type.ParseTypedBlob(
		objectType.GetType(),
		objectType.GetBlobDigest(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer repool()

	script := blob.GetStringLuaHooks()

	if script == "" {
		return err
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
		return err
	}

	return err
}

func (store *Store) tryPreCommitHooks(
	child *sku.Transacted,
	mother *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	if !options.RunHooks {
		return err
	}

	type hook struct {
		script      string
		description string
	}

	hooks := []hook{}

	var typeObject *sku.Transacted

	if typeObject, err = store.ReadOneObjectId(child.GetType()); err != nil {
		if collections.IsErrNotFound(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return err
	} else if typeObject == nil {
		return err
	}

	var blob type_blobs.Blob
	var repool interfaces.FuncRepool

	if blob, repool, _, err = store.GetTypedBlobStore().Type.ParseTypedBlob(
		typeObject.GetType(),
		typeObject.GetBlobDigest(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer repool()

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
			child,
			mother,
			options,
			typeObject,
			h.script,
			"on_pre_commit",
		); err != nil {
			err = errors.Wrapf(err, "Hook: %#v", h)
			err = errors.Wrapf(err, "Type: %q", child.GetType())

			if store.envRepo.Retry(
				"hook failed",
				"ignore error and continue?",
				err,
			) {
				// TODO fix this to properly continue past the failure
				err = nil
			} else {
				return err
			}
		}
	}

	return err
}

func (store *Store) tryPreCommitHook(
	child *sku.Transacted,
	mother *sku.Transacted,
	storeOptions sku.StoreOptions,
	selbst *sku.Transacted,
	script string,
) (err error) {
	if !storeOptions.RunHooks || !storeOptions.AddToInventoryList {
		return err
	}

	var vp sku_lua.LuaVMPoolV1

	if vp, err = store.MakeLuaVMPoolV1(selbst, script); err != nil {
		err = errors.Wrap(err)
		return err
	}

	var vm *sku_lua.LuaVMV1

	if vm, err = vp.Get(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer vp.Put(vm)

	var tt *lua.LTable

	if tt, err = vm.GetTopTableOrError(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	f := vm.GetField(tt, "on_pre_commit")

	if f.Type() != lua.LTFunction {
		return err
	}

	tableKinder := vm.TablePool.Get()
	defer vm.TablePool.Put(tableKinder)

	sku_lua.ToLuaTableV1(
		child,
		vm.LState,
		tableKinder,
	)

	var tableMutter *sku_lua.LuaTableV1

	if mother != nil {
		tableMutter = vm.TablePool.Get()
		defer vm.TablePool.Put(tableMutter)

		sku_lua.ToLuaTableV1(
			mother,
			vm.LState,
			tableMutter,
		)
	}

	vm.Push(f)
	vm.Push(tableKinder.Transacted)
	vm.Push(tableMutter.Transacted)

	if err = vm.PCall(2, 1, nil); err != nil {
		err = errors.Wrap(err)
		return err
	}

	retval := vm.LState.Get(1)
	vm.Pop(1)

	if retval.Type() != lua.LTNil {
		err = errors.ErrorWithStackf("lua error: %s", retval)
		return err
	}

	if err = sku_lua.FromLuaTableV1(child, vm.LState, tableKinder); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO add method with hook with reader
func (store *Store) tryHookWithName(
	child *sku.Transacted,
	mother *sku.Transacted,
	options sku.CommitOptions,
	self *sku.Transacted,
	script string,
	name string,
) (err error) {
	var vp sku_lua.LuaVMPoolV1

	if vp, err = store.MakeLuaVMPoolV1(self, script); err != nil {
		err = errors.Wrap(err)
		return err
	}

	var vm *sku_lua.LuaVMV1

	if vm, err = vp.Get(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err != nil {
		return err
	}

	defer vp.Put(vm)

	var tt *lua.LTable

	if tt, err = vm.GetTopTableOrError(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	f := vm.GetField(tt, name)

	if f.Type() != lua.LTFunction {
		return err
	}

	tableKinder := vm.TablePool.Get()
	defer vm.TablePool.Put(tableKinder)

	sku_lua.ToLuaTableV1(
		child,
		vm.LState,
		tableKinder,
	)

	var tableMutter *sku_lua.LuaTableV1

	if mother != nil {
		tableMutter = vm.TablePool.Get()
		defer vm.TablePool.Put(tableMutter)

		sku_lua.ToLuaTableV1(
			mother,
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
		return err
	}

	retval := vm.LState.Get(1)
	vm.Pop(1)

	if retval.Type() != lua.LTNil {
		err = errors.ErrorWithStackf("lua error: %s", retval)
		return err
	}

	if err = sku_lua.FromLuaTableV1(child, vm.LState, tableKinder); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
