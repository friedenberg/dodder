package env_lua

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
)

// TODO extract all of these components into an env_lua

type Env interface {
	MakeLuaVMPoolBuilder() *lua.VMPoolBuilder
	GetSkuFromString(lv string) (sk *sku.Transacted, err error)
}

type env struct {
	envRepo      env_repo.Env
	objectStore  sku.RepoStore
	luaSkuFormat *box_format.BoxTransacted
}

func Make(
	envRepo env_repo.Env,
	objectStore sku.RepoStore,
	luaSkuFormat *box_format.BoxTransacted,
) *env {
	return &env{
		envRepo:      envRepo,
		objectStore:  objectStore,
		luaSkuFormat: luaSkuFormat,
	}
}

func (repo *env) MakeLuaVMPoolBuilder() *lua.VMPoolBuilder {
	return (&lua.VMPoolBuilder{}).WithSearcher(repo.luaSearcher)
}

func (s *env) luaSearcher(ls *lua.LState) int {
	lv := ls.ToString(1)
	ls.Pop(1)

	var err error
	var sk *sku.Transacted

	if sk, err = s.GetSkuFromString(lv); err != nil {
		ls.Push(lua.LString(err.Error()))
		return 1
	}

	sku.GetTransactedPool().Put(sk)

	ls.Push(ls.NewFunction(s.LuaRequire))

	return 1
}

func (s *env) GetSkuFromString(lv string) (sk *sku.Transacted, err error) {
	sk = sku.GetTransactedPool().Get()

	defer func() {
		if err != nil {
			return
		}

		if err = s.objectStore.ReadOneInto(sk.GetObjectId(), sk); err != nil {
			err = errors.Wrap(err)
			return
		}
	}()

	if err = sk.ObjectId.SetOnlyNotUnknownGenre(lv); err == nil {
		return
	}

	rb := catgut.MakeRingBuffer(strings.NewReader(lv), 0)

	if _, err = s.luaSkuFormat.ReadStringFormat(
		sk,
		catgut.MakeRingBufferRuneScanner(rb),
	); err == nil {
		return
	}

	return
}

// TODO modify `package.loaded` to include variations of object id
func (s *env) LuaRequire(ls *lua.LState) int {
	// TODO handle second extra arg
	// TODO parse lv as object id / blob
	lv := ls.ToString(1)
	ls.Pop(1)

	var err error
	var sk *sku.Transacted

	if sk, err = s.GetSkuFromString(lv); err != nil {
		panic(err)
		// ls.Push(lua.LString(err.Error()))
		// return 1
	}

	defer sku.GetTransactedPool().Put(sk)

	if err = s.objectStore.ReadOneInto(sk.GetObjectId(), sk); err != nil {
		panic(err)
	}

	var ar interfaces.ReadCloseBlobIdGetter

	if ar, err = s.envRepo.GetDefaultBlobStore().BlobReader(
		sk.GetBlobId(),
	); err != nil {
		panic(err)
	}

	defer errors.DeferredCloser(&err, ar)

	var compiled *lua.FunctionProto

	if compiled, err = lua.CompileReader(ar); err != nil {
		panic(err)
	}

	ls.Push(ls.NewFunctionFromProto(compiled))

	if err = ls.PCall(0, 1, nil); err != nil {
		panic(err)
	}

	return 1
}
