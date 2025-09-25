package tag_blobs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/sku_lua"
)

func MakeLuaSelfApplyV1(
	selfOriginal *sku.Transacted,
) interfaces.FuncIter[*lua.VM] {
	if selfOriginal == nil {
		panic("self was nil")
	}

	self := selfOriginal.CloneTransacted()

	return func(vm *lua.VM) (err error) {
		selfTable := sku_lua.MakeLuaTablePoolV1(vm).Get()
		sku_lua.ToLuaTableV1(self, vm.LState, selfTable)
		vm.SetGlobal("Selbst", selfTable.Transacted)
		return err
	}
}

type LuaV1 struct {
	sku_lua.LuaVMPoolV1
}

func (a *LuaV1) GetQueryable() sku.Queryable {
	return a
}

func (a *LuaV1) Reset() {
}

func (a *LuaV1) ResetWith(b LuaV1) {
}

func (tb *LuaV1) ContainsSku(tg sku.TransactedGetter) bool {
	// lb := b.luaVMPoolBuilder.Clone().WithApply(MakeSelfApply(sk))
	vm, err := tb.Get()
	if err != nil {
		ui.Err().Printf("lua script error: %s", err)
		return false
	}

	defer tb.Put(vm)

	var t *lua.LTable

	t, err = vm.VM.GetTopTableOrError()
	if err != nil {
		ui.Err().Print(err)
		return false
	}

	// TODO safer
	f := vm.VM.GetField(t, "contains_sku").(*lua.LFunction)

	tSku := vm.TablePool.Get()
	defer vm.TablePool.Put(tSku)

	vm.VM.Push(f)

	sku_lua.ToLuaTableV1(
		tg,
		vm.VM.LState,
		tSku,
	)

	vm.VM.Push(tSku.Transacted)

	err = vm.VM.PCall(1, 1, nil)
	if err != nil {
		ui.Err().Print(err)
		return false
	}

	retval := vm.LState.Get(1)
	vm.Pop(1)

	if retval.Type() != lua.LTBool {
		ui.Err().Printf("expected bool but got %s", retval.Type())
		return false
	}

	return bool(retval.(lua.LBool))
}
