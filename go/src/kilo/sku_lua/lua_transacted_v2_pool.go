package sku_lua

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type LuaVMV2 struct {
	lua.LValue
	*lua.VM
	TablePool LuaTablePoolV2
	Selbst    *sku.Transacted
}

func PushTopFuncV2(
	luaVMPool LuaVMPoolV2,
	args []string,
) (vm *LuaVMV2, argsOut []string, err error) {
	if vm, err = luaVMPool.Get(); err != nil {
		err = errors.Wrap(err)
		return vm, argsOut, err
	}

	vm.LValue = vm.Top

	var luaFunction *lua.LFunction

	if luaFunction, argsOut, err = vm.GetTopFunctionOrFunctionNamedError(
		args,
	); err != nil {
		err = errors.Wrap(err)
		return vm, argsOut, err
	}

	vm.Push(luaFunction)

	return vm, argsOut, err
}

type (
	LuaVMPoolV2    interfaces.PoolWithErrorsPtr[LuaVMV2, *LuaVMV2]
	LuaTablePoolV2 = interfaces.Pool[LuaTableV2, *LuaTableV2]
)

func MakeLuaVMPoolV2(luaVMPool *lua.VMPool, self *sku.Transacted) LuaVMPoolV2 {
	return pool.MakeWithError(
		func() (out *LuaVMV2, err error) {
			var vm *lua.VM

			if vm, err = luaVMPool.PoolWithErrorsPtr.Get(); err != nil {
				err = errors.Wrap(err)
				return out, err
			}

			out = &LuaVMV2{
				VM:        vm,
				TablePool: MakeLuaTablePoolV2(vm),
				Selbst:    self,
			}

			return out, err
		},
		nil,
	)
}

func MakeLuaTablePoolV2(vm *lua.VM) LuaTablePoolV2 {
	return pool.Make(
		func() (t *LuaTableV2) {
			t = &LuaTableV2{
				Transacted:   vm.Pool.Get(),
				Tags:         vm.Pool.Get(),
				TagsImplicit: vm.Pool.Get(),
			}

			vm.SetField(t.Transacted, "Tags", t.Tags)
			vm.SetField(t.Transacted, "TagsImplicit", t.TagsImplicit)

			return t
		},
		func(t *LuaTableV2) {
			lua.ClearTable(vm.LState, t.Tags)
			lua.ClearTable(vm.LState, t.TagsImplicit)
		},
	)
}
