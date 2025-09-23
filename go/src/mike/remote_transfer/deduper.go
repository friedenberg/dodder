package remote_transfer

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
)

type deduper struct {
	formatId   string
	lookupLock *sync.RWMutex
	lookup     map[string]struct{}
	id         markl.Id
}

func (deduper *deduper) initialize(
	options repo.ImporterOptions,
	envRepo env_repo.Env,
) {
	if options.DedupingFormatId != "" {
		deduper.formatId = options.DedupingFormatId
		deduper.lookupLock = &sync.RWMutex{}
		deduper.lookup = make(map[string]struct{})
	}
}

func (deduper *deduper) shouldCommit(object *sku.Transacted) (err error) {
	if deduper.lookup == nil {
		return err
	}

	objectDigestWriteMap := object.GetDigestWriteMapWithMerkle()

	var id interfaces.MutableMarklId

	{
		var hasDigest bool

		if id, hasDigest = objectDigestWriteMap[deduper.formatId]; !hasDigest {
			err = errors.Errorf(
				"object does not have digest for format id: %q",
				deduper.formatId,
			)

			return err
		}
	}

	bites := id.GetBytes()

	deduper.lookupLock.RLock()
	if _, exists := deduper.lookup[string(bites)]; exists {
		deduper.lookupLock.RUnlock()
		return ErrSkipped
	}

	deduper.lookupLock.RUnlock()

	deduper.lookupLock.Lock()
	deduper.lookup[string(bites)] = struct{}{}
	deduper.lookupLock.Unlock()

	return err
}
