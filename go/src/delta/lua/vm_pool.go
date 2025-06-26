package lua

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/pool2"
	lua "github.com/yuin/gopher-lua"
)

type VMPool struct {
	interfaces.Pool2[VM, *VM]
	Require  LGFunction
	Searcher LGFunction
	compiled *lua.FunctionProto
}

func (sp *VMPool) PrepareVM(
	vm *VM,
	apply interfaces.FuncIter[*VM],
) (err error) {
	vm.Pool = pool.MakePool(
		func() (t *lua.LTable) {
			t = vm.NewTable()
			return
		},
		func(t *lua.LTable) {
			ClearTable(vm.LState, t)
		},
	)

	if sp.Require != nil {
		vm.PreloadModule("zit", func(s *lua.LState) int {
			// register functions to the table
			mod := s.SetFuncs(s.NewTable(), map[string]lua.LGFunction{
				"require": sp.Require,
			})

			s.Push(mod)

			return 1
		})

		tableZit := vm.Pool.Get()
		vm.SetField(tableZit, "require", vm.NewFunction(sp.Require))
		vm.SetGlobal("zit", tableZit)
	}

	if sp.Searcher != nil {
		packageTable := vm.GetGlobal("package").(*LTable)

		if true { // lua <= 5.1
			loaderTable := vm.GetField(packageTable, "loaders").(*LTable)
			loaderTable.Insert(1, vm.NewFunction(sp.Searcher))
		} else {
			searcherTable := vm.Pool.Get()
			packageTable.Insert(1, searcherTable)
			searcherTable.Insert(1, vm.NewFunction(sp.Searcher))
		}
	}

	if apply != nil {
		if err = apply(vm); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	lfunc := vm.NewFunctionFromProto(sp.compiled)
	vm.Push(lfunc)

	if err = vm.PCall(0, 1, nil); err != nil {
		err = errors.Wrap(err)
		return
	}

	vm.Top = vm.LState.Get(1)
	vm.Pop(1)

	return
}

func (sp *VMPool) SetReader(
	reader io.Reader,
	apply interfaces.FuncIter[*VM],
) (err error) {
	var compiled *FunctionProto

	if compiled, err = CompileReader(reader); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = sp.SetCompiled(compiled, apply); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (sp *VMPool) SetCompiled(
	compiled *FunctionProto,
	apply interfaces.FuncIter[*VM],
) (err error) {
	sp.compiled = compiled

	sp.Pool2 = pool2.MakePool(
		func() (vm *VM, err error) {
			vm = &VM{
				LState: lua.NewState(),
			}

			if err = sp.PrepareVM(vm, apply); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		},
		func(vm *VM) {
			vm.SetTop(0)
		},
	)

	return
}
