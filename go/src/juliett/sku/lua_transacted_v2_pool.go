package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
)

type LuaVMV2 struct {
	lua.LValue
	*lua.VM
	TablePool LuaTablePoolV2
	Selbst    *Transacted
}

func PushTopFuncV2(
	luaVMPool LuaVMPoolV2,
	args []string,
) (vm *LuaVMV2, argsOut []string, err error) {
	if vm, err = luaVMPool.Get(); err != nil {
		err = errors.Wrap(err)
		return
	}

	vm.LValue = vm.Top

	var luaFunction *lua.LFunction

	if luaFunction, argsOut, err = vm.GetTopFunctionOrFunctionNamedError(
		args,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	vm.Push(luaFunction)

	return
}

type (
	LuaVMPoolV2    interfaces.PoolWithErrorsPtr[LuaVMV2, *LuaVMV2]
	LuaTablePoolV2 = interfaces.Pool[LuaTableV2, *LuaTableV2]
)

func MakeLuaVMPoolV2(luaVMPool *lua.VMPool, self *Transacted) LuaVMPoolV2 {
	return pool.MakeWithError(
		func() (out *LuaVMV2, err error) {
			var vm *lua.VM

			if vm, err = luaVMPool.PoolWithErrorsPtr.Get(); err != nil {
				err = errors.Wrap(err)
				return
			}

			out = &LuaVMV2{
				VM:        vm,
				TablePool: MakeLuaTablePoolV2(vm),
				Selbst:    self,
			}

			return
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

			return
		},
		func(t *LuaTableV2) {
			lua.ClearTable(vm.LState, t.Tags)
			lua.ClearTable(vm.LState, t.TagsImplicit)
		},
	)
}
