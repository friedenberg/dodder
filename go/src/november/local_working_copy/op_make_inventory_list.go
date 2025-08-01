package local_working_copy

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
)

func (local *Repo) MakeInventoryList(
	queryGroup *query.Query,
) (list *sku.ListTransacted, err error) {
	list = sku.MakeListTransacted()

	var l sync.Mutex

	if err = local.GetStore().QueryTransacted(
		queryGroup,
		func(sk *sku.Transacted) (err error) {
			l.Lock()
			defer l.Unlock()

			list.Add(sk.CloneTransacted())

			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
