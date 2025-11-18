package sku

import (
	"encoding/gob"

	"code.linenisgreat.com/dodder/go/src/_/external_state"
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

func init() {
	gob.Register(Transacted{})
}

type (
	Config interface {
		interfaces.Config
		ids.InlineTypeChecker // TODO move out of konfig entirely
	}

	ObjectProbeIndex interface {
		ReadOneObjectId(interfaces.ObjectId, *Transacted) error
	}

	TransactedGetter interface {
		GetSku() *Transacted
	}

	ObjectWithList struct {
		Object, List *Transacted
	}

	ExternalLike interface {
		ids.ObjectIdGetter
		interfaces.Stringer
		TransactedGetter
		ExternalLikeGetter
		GetExternalState() external_state.State
		ExternalObjectIdGetter
		GetRepoId() ids.RepoId
	}

	ExternalLikeGetter interface {
		GetSkuExternal() *Transacted
	}

	FSItemReadWriter interface {
		ReadFSItemFromExternal(TransactedGetter) (*FSItem, error)
		WriteFSItemToExternal(*FSItem, TransactedGetter) (err error)
	}

	OneReader interface {
		ReadTransactedFromObjectId(
			k1 interfaces.ObjectId,
		) (sk1 *Transacted, err error)
	}

	BlobSaver interface {
		SaveBlob(ExternalLike) (err error)
	}
)
