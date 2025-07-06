package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/mike/importer"
)

func (local *Repo) ImportList(
	list *sku.List,
	i sku.Importer,
) (err error) {
	local.Must(local.Lock)

	if err = local.GetInventoryListStore().ImportList(
		list,
		i,
	); err != nil {
		if !errors.Is(err, importer.ErrNeedsMerge) {
			err = errors.Wrap(err)
			return
		}
	}

	local.Must(local.Unlock)

	return
}
