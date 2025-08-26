package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/pool2"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
)

type LuaVMV1 struct {
	lua.LValue
	*lua.VM
	TablePool LuaTablePoolV1
	Selbst    *Transacted
}

func PushTopFuncV1(
	lvm LuaVMPoolV1,
	args []string,
) (vm *LuaVMV1, argsOut []string, err error) {
	if vm, err = lvm.Get(); err != nil {
		err = errors.Wrap(err)
		return
	}

	vm.LValue = vm.Top

	var f *lua.LFunction

	if f, argsOut, err = vm.GetTopFunctionOrFunctionNamedError(
		args,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	vm.Push(f)

	return
}

type (
	LuaVMPoolV1    interfaces.PoolWithErrorsPtr[LuaVMV1, *LuaVMV1]
	LuaTablePoolV1 = interfaces.Pool[LuaTableV1, *LuaTableV1]
)

func MakeLuaVMPoolV1(lvp *lua.VMPool, selbst *Transacted) LuaVMPoolV1 {
	return pool2.MakePool(
		func() (out *LuaVMV1, err error) {
			var vm *lua.VM

			if vm, err = lvp.PoolWithErrorsPtr.Get(); err != nil {
				err = errors.Wrap(err)
				return
			}

			out = &LuaVMV1{
				VM:        vm,
				TablePool: MakeLuaTablePoolV1(vm),
				Selbst:    selbst,
			}

			return
		},
		nil,
	)
}

func MakeLuaTablePoolV1(vm *lua.VM) LuaTablePoolV1 {
	return pool.Make(
		func() (t *LuaTableV1) {
			t = &LuaTableV1{
				Transacted:   vm.Pool.Get(),
				Tags:         vm.Pool.Get(),
				TagsImplicit: vm.Pool.Get(),
			}

			vm.SetField(t.Transacted, "Etiketten", t.Tags)
			vm.SetField(t.Transacted, "EtikettenImplicit", t.TagsImplicit)

			return
		},
		func(t *LuaTableV1) {
			lua.ClearTable(vm.LState, t.Tags)
			lua.ClearTable(vm.LState, t.TagsImplicit)
		},
	)
}
