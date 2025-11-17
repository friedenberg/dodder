package store_workspace

import "code.linenisgreat.com/dodder/go/src/foxtrot/ids"

type (
	Store interface {
		GetObjectIdsForString(string) ([]ids.ExternalObjectIdLike, error)
	}

	StoreGetter interface {
		GetWorkspaceStoreForQuery(ids.RepoId) (Store, bool)
	}
)
