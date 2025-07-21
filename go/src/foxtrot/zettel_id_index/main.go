package zettel_id_index

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
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
	config genesis_configs.ConfigPublic,
	configCli repo_config_cli.Blob,
	directory interfaces.Directory,
	cacheIOFactory interfaces.CacheIOFactory,
) (i Index, err error) {
	if store_version.GreaterOrEqual(
		config.GetStoreVersion(),
		store_version.V1,
	) && false {
		ui.TodoP3("investigate using bitsets")
		if i, err = hinweis_index_v1.MakeIndex(
			configCli,
			directory,
			cacheIOFactory,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	} else {
		if i, err = hinweis_index_v0.MakeIndex(
			configCli,
			directory,
			cacheIOFactory,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
