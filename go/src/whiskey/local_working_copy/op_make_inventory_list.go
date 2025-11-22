package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
	"code.linenisgreat.com/dodder/go/src/papa/queries"
)

func (local *Repo) MakeInventoryList(
	query *queries.Query,
) (list *sku.HeapTransactedTai, err error) {
	list = sku.MakeListTransacted()

	if err = local.GetStore().QueryTransacted(
		query,
		quiter.MakeSyncSerializer(
			func(object *sku.Transacted) (err error) {
				return list.Add(object.CloneTransacted())
			},
		),
	); err != nil {
		err = errors.Wrap(err)
		return list, err
	}

	return list, err
}
