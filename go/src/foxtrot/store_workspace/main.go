package store_workspace

import (
	"code.linenisgreat.com/dodder/go/src/alfa/domain_interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type (
	Store interface {
		GetObjectIdsForString(string) ([]domain_interfaces.ExternalObjectId, error)
	}

	StoreGetter interface {
		GetWorkspaceStoreForQuery(ids.RepoId) (Store, bool)
	}
)
