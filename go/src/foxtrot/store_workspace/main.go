package store_workspace

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type (
	Store interface {
		GetObjectIdsForString(string) ([]interfaces.ExternalObjectId, error)
	}

	StoreGetter interface {
		GetWorkspaceStoreForQuery(ids.RepoId) (Store, bool)
	}
)
