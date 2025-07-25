package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type LuaTableV2 struct {
	Transacted *lua.LTable

	// TODO transition to single Tags table with Tag objects that reflect
	// tag_paths.PathWithType
	Tags         *lua.LTable
	TagsImplicit *lua.LTable
}

func ToLuaTableV2(tg TransactedGetter, l *lua.LState, t *LuaTableV2) {
	o := tg.GetSku()

	l.SetField(t.Transacted, "Genre", lua.LString(o.GetGenre().String()))
	l.SetField(t.Transacted, "ObjectId", lua.LString(o.GetObjectId().String()))
	l.SetField(t.Transacted, "Type", lua.LString(o.GetType().String()))

	tags := t.Tags

	for e := range o.Metadata.GetTags().AllPtr() {
		l.SetField(tags, e.String(), lua.LBool(true))
	}

	tags = t.TagsImplicit

	for e := range o.Metadata.Cache.GetImplicitTags().AllPtr() {
		l.SetField(tags, e.String(), lua.LBool(true))
	}
}

func FromLuaTableV2(o *Transacted, l *lua.LState, lt *LuaTableV2) (err error) {
	t := lt.Transacted

	g := genres.MakeOrUnknown(l.GetField(t, "Genre").String())

	o.ObjectId.SetGenre(g)
	k := l.GetField(t, "ObjectId").String()

	if k != "" {
		if err = o.ObjectId.Set(k); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	et := l.GetField(t, "Tags")
	ets, ok := et.(*lua.LTable)

	if !ok {
		err = errors.ErrorWithStackf("expected table but got %T", et)
		return
	}

	o.Metadata.SetTags(nil)

	ets.ForEach(
		func(key, value lua.LValue) {
			var e ids.Tag

			if err = e.Set(key.String()); err != nil {
				err = errors.Wrap(err)
				panic(err)
			}

			errors.PanicIfError(o.Metadata.AddTagPtr(&e))
		},
	)

	// TODO Description
	// TODO Type
	// TODO Tai
	// TODO Blob
	// TODO Cache

	return
}
