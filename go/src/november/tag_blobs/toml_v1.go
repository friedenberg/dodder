package tag_blobs

import (
	"code.linenisgreat.com/dodder/go/src/bravo/lua"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
	"code.linenisgreat.com/dodder/go/src/mike/sku_lua"
)

type TomlV1 struct {
	sku_lua.LuaVMPoolV1
	Filter string `toml:"filter"`
}

func (a *TomlV1) Reset() {
	a.Filter = ""
}

func (a *TomlV1) ResetWith(b TomlV1) {
	a.Filter = b.Filter
}

func (tb *TomlV1) ContainsSku(tg sku.TransactedGetter) bool {
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
