package store

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/sku_lua"
	"code.linenisgreat.com/dodder/go/src/kilo/tag_blobs"
)

func (store *Store) MakeLuaVMPoolV1WithSku(
	sk *sku.Transacted,
) (lvp sku_lua.LuaVMPoolV1, err error) {
	if sk.GetType().String() != "lua" {
		err = errors.ErrorWithStackf(
			"unsupported typ: %s, Sku: %s",
			sk.GetType(),
			sk,
		)
		return
	}

	var readCloser interfaces.BlobReader

	if readCloser, err = store.GetEnvRepo().GetDefaultBlobStore().MakeBlobReader(sk.GetBlobDigest()); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, readCloser)

	if lvp, err = store.MakeLuaVMPoolWithReader(sk, readCloser); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) MakeLuaVMPoolV1(
	self *sku.Transacted,
	script string,
) (vp sku_lua.LuaVMPoolV1, err error) {
	b := store.envLua.MakeLuaVMPoolBuilder().
		WithScript(script).
		WithApply(tag_blobs.MakeLuaSelfApplyV1(self))

	var lvmp *lua.VMPool

	if lvmp, err = b.Build(); err != nil {
		err = errors.Wrap(err)
		return
	}

	vp = sku_lua.MakeLuaVMPoolV1(lvmp, self)

	return
}

func (store *Store) MakeLuaVMPoolWithReader(
	selbst *sku.Transacted,
	r io.Reader,
) (vp sku_lua.LuaVMPoolV1, err error) {
	b := store.envLua.MakeLuaVMPoolBuilder().
		WithReader(r).
		WithApply(tag_blobs.MakeLuaSelfApplyV1(selbst))

	var lvmp *lua.VMPool

	if lvmp, err = b.Build(); err != nil {
		err = errors.Wrap(err)
		return
	}

	vp = sku_lua.MakeLuaVMPoolV1(lvmp, selbst)

	return
}
