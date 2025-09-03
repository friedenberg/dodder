package importer

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type deduper struct {
	format object_inventory_format.Format
	lookup map[string]struct{}
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

		deduper.lookup = make(map[string]struct{})
	}
}

func (deduper *deduper) shouldCommit(object *sku.Transacted) error {
	return nil
}
