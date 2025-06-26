package zettel_id_index

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	hinweis_index_v0 "code.linenisgreat.com/dodder/go/src/foxtrot/zettel_id_index/v0"
	hinweis_index_v1 "code.linenisgreat.com/dodder/go/src/foxtrot/zettel_id_index/v1"
)

type Index interface {
	errors.Flusher
	CreateZettelId() (*ids.ZettelId, error)
	interfaces.ResetterWithError
	AddZettelId(interfaces.ObjectId) error
	PeekZettelIds(int) ([]*ids.ZettelId, error)
}

func MakeIndex(
	k interfaces.Config,
	s interfaces.Directory,
	su interfaces.CacheIOFactory,
) (i Index, err error) {
	if store_version.GreaterOrEqual(
		k.GetStoreVersion(),
		store_version.V1,
	) && false {
		ui.TodoP3("investigate using bitsets")
		if i, err = hinweis_index_v1.MakeIndex(
			k,
			s,
			su,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	} else {
		if i, err = hinweis_index_v0.MakeIndex(
			k,
			s,
			su,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
