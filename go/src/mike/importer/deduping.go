package importer

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type deduper struct {
	format     object_inventory_format.Format
	lookupLock *sync.RWMutex
	lookup     map[string]struct{}
}

func (deduper *deduper) initialize(
	options sku.ImporterOptions,
	envRepo env_repo.Env,
) {
	if options.DedupingFormatId != "" {
		var err error
		if deduper.format, err = object_inventory_format.FormatForMarklFormatIdError(
			options.DedupingFormatId,
		); err != nil {
			errors.ContextCancelWithBadRequestf(
				envRepo,
				"format id for deduping not found: %q",
				options.DedupingFormatId,
			)
		}

		deduper.lookupLock = &sync.RWMutex{}
		deduper.lookup = make(map[string]struct{})
	}
}

func (deduper *deduper) shouldCommit(object *sku.Transacted) (err error) {
	if deduper.lookup == nil {
		return
	}

	id, repool := markl.HashTypeSha256.GetBlobId()
	defer repool()

	if err = object.CalculateDigest(
		object_inventory_format.GetDigestForContext,
		deduper.format,
		id,
	); err != nil {
		err = errors.Wrap(err)
		return
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

	return
}
