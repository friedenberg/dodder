package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/mike/importer"
)

func (local *Repo) ImportSeq(
	seq interfaces.SeqError[*sku.Transacted],
	importerr sku.Importer,
) (err error) {
	local.Must(errors.MakeFuncContextFromFuncErr(local.Lock))

	if err = local.GetInventoryListStore().ImportSeq(
		seq,
		importerr,
	); err != nil {
		if !errors.Is(err, importer.ErrNeedsMerge) {
			err = errors.Wrap(err)
			return
		}
	}

	local.Must(errors.MakeFuncContextFromFuncErr(local.Unlock))

	return
}
