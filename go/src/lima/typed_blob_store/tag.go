package typed_blob_store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/sku_lua"
	"code.linenisgreat.com/dodder/go/src/kilo/tag_blobs"
	"code.linenisgreat.com/dodder/go/src/lima/env_lua"
)

type Tag struct {
	envRepo env_repo.Env
	envLua  env_lua.Env
	toml_v0 TypedStore[tag_blobs.V0, *tag_blobs.V0]
	toml_v1 TypedStore[tag_blobs.TomlV1, *tag_blobs.TomlV1]
	lua_v1  TypedStore[tag_blobs.LuaV1, *tag_blobs.LuaV1]
	lua_v2  TypedStore[tag_blobs.LuaV2, *tag_blobs.LuaV2]
}

func MakeTagStore(
	envRepo env_repo.Env,
	envLua env_lua.Env,
) Tag {
	return Tag{
		envRepo: envRepo,
		envLua:  envLua,
		toml_v0: MakeBlobStore(
			envRepo,
			MakeBlobFormat(
				MakeTomlDecoderIgnoreTomlErrors[tag_blobs.V0](
					envRepo.GetDefaultBlobStore(),
				),
				TomlBlobEncoder[tag_blobs.V0, *tag_blobs.V0]{},
				envRepo.GetDefaultBlobStore(),
			),
			func(a *tag_blobs.V0) {
				a.Reset()
			},
		),
		toml_v1: MakeBlobStore(
			envRepo,
			MakeBlobFormat(
				MakeTomlDecoderIgnoreTomlErrors[tag_blobs.TomlV1](
					envRepo.GetDefaultBlobStore(),
				),
				TomlBlobEncoder[tag_blobs.TomlV1, *tag_blobs.TomlV1]{},
				envRepo.GetDefaultBlobStore(),
			),
			func(a *tag_blobs.TomlV1) {
				a.Reset()
			},
		),
		lua_v1: MakeBlobStore(
			envRepo,
			MakeBlobFormat[tag_blobs.LuaV1](
				nil,
				nil,
				envRepo.GetDefaultBlobStore(),
			),
			func(a *tag_blobs.LuaV1) {
			},
		),
		lua_v2: MakeBlobStore(
			envRepo,
			MakeBlobFormat[tag_blobs.LuaV2](
				nil,
				nil,
				envRepo.GetDefaultBlobStore(),
			),
			func(a *tag_blobs.LuaV2) {
			},
		),
	}
}

// TODO check repool funcs
func (store Tag) GetBlob(
	object *sku.Transacted,
) (blobGeneric tag_blobs.Blob, repool interfaces.FuncRepool, err error) {
	tipe := object.GetType()
	blobDigest := object.GetBlobDigest()

	switch tipe.String() {
	case "", ids.TypeTomlTagV0:
		if blobGeneric, repool, err = store.toml_v0.GetBlob(blobDigest); err != nil {
			err = errors.Wrap(err)
			return
		}

	case ids.TypeTomlTagV1:
		var blob *tag_blobs.TomlV1

		if blob, repool, err = store.toml_v1.GetBlob(blobDigest); err != nil {
			err = errors.Wrap(err)
			return
		}

		luaVMPoolBuilder := store.envLua.MakeLuaVMPoolBuilder().WithApply(
			tag_blobs.MakeLuaSelfApplyV1(object),
		)

		var luaVMPool *lua.VMPool

		luaVMPoolBuilder.WithScript(blob.Filter)

		if luaVMPool, err = luaVMPoolBuilder.Build(); err != nil {
			err = errors.Wrap(err)
			return
		}

		blob.LuaVMPoolV1 = sku_lua.MakeLuaVMPoolV1(luaVMPool, nil)
		blobGeneric = blob

	case ids.TypeLuaTagV1:
		// TODO try to repool things here

		var readCloser interfaces.ReadCloseMarklIdGetter

		if readCloser, err = store.envRepo.GetDefaultBlobStore().MakeBlobReader(
			blobDigest,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer errors.DeferredCloser(&err, readCloser)

		luaVMPoolBuilder := store.envLua.MakeLuaVMPoolBuilder().WithApply(
			tag_blobs.MakeLuaSelfApplyV1(object),
		)

		var luaVMPool *lua.VMPool

		luaVMPoolBuilder.WithReader(readCloser)

		if luaVMPool, err = luaVMPoolBuilder.Build(); err != nil {
			err = errors.Wrap(err)
			return
		}

		blobGeneric = &tag_blobs.LuaV1{
			LuaVMPoolV1: sku_lua.MakeLuaVMPoolV1(luaVMPool, nil),
		}

	case ids.TypeLuaTagV2:
		// TODO try to repool things here

		var readCloser interfaces.ReadCloseMarklIdGetter

		if readCloser, err = store.envRepo.GetDefaultBlobStore().MakeBlobReader(blobDigest); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer errors.DeferredCloser(&err, readCloser)

		luaVMPoolBUilder := store.envLua.MakeLuaVMPoolBuilder().WithApply(
			tag_blobs.MakeLuaSelfApplyV2(object),
		)

		var luaVMPool *lua.VMPool

		luaVMPoolBUilder.WithReader(readCloser)

		if luaVMPool, err = luaVMPoolBUilder.Build(); err != nil {
			err = errors.Wrap(err)
			return
		}

		blobGeneric = &tag_blobs.LuaV2{
			LuaVMPoolV2: sku_lua.MakeLuaVMPoolV2(luaVMPool, nil),
		}
	}

	return
}
