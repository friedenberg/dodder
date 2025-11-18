package typed_blob_store

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/lua"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/kilo/env_repo"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
	"code.linenisgreat.com/dodder/go/src/mike/sku_lua"
	"code.linenisgreat.com/dodder/go/src/november/env_lua"
	"code.linenisgreat.com/dodder/go/src/november/tag_blobs"
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
	blobId := object.GetBlobDigest()

	switch tipe.String() {
	case "", ids.TypeTomlTagV0:
		if blobGeneric, repool, err = store.toml_v0.GetBlob(blobId); err != nil {
			err = errors.Wrap(err)
			return blobGeneric, repool, err
		}

	case ids.TypeTomlTagV1:
		var blob *tag_blobs.TomlV1

		if blob, repool, err = store.toml_v1.GetBlob(blobId); err != nil {
			err = errors.Wrap(err)
			return blobGeneric, repool, err
		}

		luaVMPoolBuilder := store.envLua.MakeLuaVMPoolBuilder().WithApply(
			tag_blobs.MakeLuaSelfApplyV1(object),
		)

		var luaVMPool *lua.VMPool

		luaVMPoolBuilder.WithScript(blob.Filter)

		if luaVMPool, err = luaVMPoolBuilder.Build(); err != nil {
			err = errors.Wrap(err)
			return blobGeneric, repool, err
		}

		blob.LuaVMPoolV1 = sku_lua.MakeLuaVMPoolV1(luaVMPool, nil)
		blobGeneric = blob

	case ids.TypeLuaTagV1:
		// TODO try to repool things here

		var readCloser interfaces.BlobReader

		if readCloser, err = store.envRepo.GetDefaultBlobStore().MakeBlobReader(
			blobId,
		); err != nil {
			err = errors.Wrap(err)
			return blobGeneric, repool, err
		}

		defer errors.DeferredCloser(&err, readCloser)

		luaVMPoolBuilder := store.envLua.MakeLuaVMPoolBuilder().WithApply(
			tag_blobs.MakeLuaSelfApplyV1(object),
		)

		var luaVMPool *lua.VMPool

		luaVMPoolBuilder.WithReader(readCloser)

		if luaVMPool, err = luaVMPoolBuilder.Build(); err != nil {
			err = errors.Wrap(err)
			return blobGeneric, repool, err
		}

		blobGeneric = &tag_blobs.LuaV1{
			LuaVMPoolV1: sku_lua.MakeLuaVMPoolV1(luaVMPool, nil),
		}

	case ids.TypeLuaTagV2:
		// TODO try to repool things here

		var readCloser interfaces.BlobReader

		if readCloser, err = store.envRepo.GetDefaultBlobStore().MakeBlobReader(blobId); err != nil {
			err = errors.Wrap(err)
			return blobGeneric, repool, err
		}

		defer errors.DeferredCloser(&err, readCloser)

		luaVMPoolBUilder := store.envLua.MakeLuaVMPoolBuilder().WithApply(
			tag_blobs.MakeLuaSelfApplyV2(object),
		)

		var luaVMPool *lua.VMPool

		luaVMPoolBUilder.WithReader(readCloser)

		if luaVMPool, err = luaVMPoolBUilder.Build(); err != nil {
			err = errors.Wrap(err)
			return blobGeneric, repool, err
		}

		blobGeneric = &tag_blobs.LuaV2{
			LuaVMPoolV2: sku_lua.MakeLuaVMPoolV2(luaVMPool, nil),
		}
	}

	return blobGeneric, repool, err
}
