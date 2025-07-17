package local_working_copy

import "code.linenisgreat.com/dodder/go/src/alfa/errors"

func (local *Repo) Reindex() {
	local.Must(errors.MakeFuncContextFromFuncErr(local.Lock))
	local.Must(errors.MakeFuncContextFromFuncErr(local.config.Reset))
	local.Must(errors.MakeFuncContextFromFuncErr(local.GetStore().Reindex))
	local.Must(errors.MakeFuncContextFromFuncErr(local.Unlock))
}
