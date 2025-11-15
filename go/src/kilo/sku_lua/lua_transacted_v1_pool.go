package sku_lua

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type LuaVMV1 struct {
	lua.LValue
	*lua.VM
	TablePool LuaTablePoolV1
	Selbst    *sku.Transacted
}

func PushTopFuncV1(
	lvm LuaVMPoolV1,
	args []string,
) (vm *LuaVMV1, argsOut []string, err error) {
	if vm, err = lvm.Get(); err != nil {
		err = errors.Wrap(err)
		return vm, argsOut, err
	}

	vm.LValue = vm.Top

	var f *lua.LFunction

	if f, argsOut, err = vm.GetTopFunctionOrFunctionNamedError(
		args,
	); err != nil {
		err = errors.Wrap(err)
		return vm, argsOut, err
	}

	vm.Push(f)

	return vm, argsOut, err
}

type (
	LuaVMPoolV1    interfaces.PoolWithErrorsPtr[LuaVMV1, *LuaVMV1]
	LuaTablePoolV1 = interfaces.Pool[LuaTableV1, *LuaTableV1]
)

func MakeLuaVMPoolV1(vmPool *lua.VMPool, self *sku.Transacted) LuaVMPoolV1 {
	return pool.MakeWithError(
		func() (out *LuaVMV1, err error) {
			var vm *lua.VM

			if vm, err = vmPool.PoolWithErrorsPtr.Get(); err != nil {
				err = errors.Wrap(err)
				return out, err
			}

			out = &LuaVMV1{
				VM:        vm,
				TablePool: MakeLuaTablePoolV1(vm),
				Selbst:    self,
			}

			return out, err
		},
		nil,
	)
}

func MakeLuaTablePoolV1(vm *lua.VM) LuaTablePoolV1 {
	return pool.Make(
		func() (table *LuaTableV1) {
			table = &LuaTableV1{
				Transacted:   vm.Pool.Get(),
				Tags:         vm.Pool.Get(),
				TagsImplicit: vm.Pool.Get(),
			}

			vm.SetField(table.Transacted, "Etiketten", table.Tags)
			vm.SetField(
				table.Transacted,
				"EtikettenImplicit",
				table.TagsImplicit,
			)

			return table
		},
		func(t *LuaTableV1) {
			lua.ClearTable(vm.LState, t.Tags)
			lua.ClearTable(vm.LState, t.TagsImplicit)
		},
	)
}
