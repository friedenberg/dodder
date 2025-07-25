package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type LuaTableV1 struct {
	Transacted   *lua.LTable
	Tags         *lua.LTable
	TagsImplicit *lua.LTable
}

func ToLuaTableV1(tg TransactedGetter, l *lua.LState, t *LuaTableV1) {
	o := tg.GetSku()

	l.SetField(t.Transacted, "Gattung", lua.LString(o.GetGenre().String()))
	l.SetField(t.Transacted, "Kennung", lua.LString(o.GetObjectId().String()))
	l.SetField(t.Transacted, "Typ", lua.LString(o.GetType().String()))

	tags := t.Tags

	for e := range o.Metadata.GetTags().AllPtr() {
		l.SetField(tags, e.String(), lua.LBool(true))
	}

	tags = t.TagsImplicit

	for e := range o.Metadata.Cache.GetImplicitTags().AllPtr() {
		l.SetField(tags, e.String(), lua.LBool(true))
	}
}

func FromLuaTableV1(o *Transacted, l *lua.LState, lt *LuaTableV1) (err error) {
	t := lt.Transacted

	g := genres.MakeOrUnknown(l.GetField(t, "Gattung").String())

	o.ObjectId.SetGenre(g)
	k := l.GetField(t, "Kennung").String()

	if k != "" {
		if err = o.ObjectId.Set(k); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	et := l.GetField(t, "Etiketten")
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

	// TODO Bezeichnung
	// TODO Typ
	// TODO Tai
	// TODO Blob
	// TODO Verzeichnisse

	return
}
