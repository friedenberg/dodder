package sku_lua

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type LuaTableV1 struct {
	Transacted   *lua.LTable
	Tags         *lua.LTable
	TagsImplicit *lua.LTable
}

func ToLuaTableV1(
	tg sku.TransactedGetter,
	luaState *lua.LState,
	luaTable *LuaTableV1,
) {
	object := tg.GetSku()

	luaState.SetField(
		luaTable.Transacted,
		"Gattung",
		lua.LString(object.GetGenre().String()),
	)
	luaState.SetField(
		luaTable.Transacted,
		"Kennung",
		lua.LString(object.GetObjectId().String()),
	)
	luaState.SetField(
		luaTable.Transacted,
		"Typ",
		lua.LString(object.GetType().String()),
	)

	tags := luaTable.Tags

	for tag := range object.Metadata.GetTags().AllPtr() {
		luaState.SetField(tags, tag.String(), lua.LBool(true))
	}

	tags = luaTable.TagsImplicit

	for tag := range object.Metadata.Index.GetImplicitTags().AllPtr() {
		luaState.SetField(tags, tag.String(), lua.LBool(true))
	}
}

func FromLuaTableV1(
	object *sku.Transacted,
	luaState *lua.LState,
	luaTable *LuaTableV1,
) (err error) {
	transacted := luaTable.Transacted

	genre := genres.MakeOrUnknown(
		luaState.GetField(transacted, "Gattung").String(),
	)

	object.ObjectId.SetGenre(genre)
	id := luaState.GetField(transacted, "Kennung").String()

	if id != "" {
		if err = object.ObjectId.Set(id); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	tags := luaState.GetField(transacted, "Etiketten")
	tagsTable, ok := tags.(*lua.LTable)

	if !ok {
		err = errors.ErrorWithStackf("expected table but got %T", tags)
		return err
	}

	object.Metadata.SetTags(nil)

	tagsTable.ForEach(
		func(key, value lua.LValue) {
			var tag ids.Tag

			if err = tag.Set(key.String()); err != nil {
				err = errors.Wrap(err)
				panic(err)
			}

			errors.PanicIfError(object.Metadata.AddTagPtr(&tag))
		},
	)

	// TODO Bezeichnung
	// TODO Typ
	// TODO Tai
	// TODO Blob
	// TODO Verzeichnisse

	return err
}
