package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/tango/repo"
)

func (local *Repo) ImportSeq(
	seq interfaces.SeqError[*sku.Transacted],
	importer repo.Importer,
) (err error) {
	return importer.ImportSeq(
		local,
		local,
		local,
		seq,
	)
}
