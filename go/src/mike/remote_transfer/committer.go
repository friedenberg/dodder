package remote_transfer

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type committer struct {
	options     sku.ImporterOptions
	storeObject sku.RepoStore
	deduper     deduper
}

func (committer *committer) initialize(
	options sku.ImporterOptions,
	envRepo env_repo.Env,
	storeObject sku.RepoStore,
) {
	committer.options = options
	committer.storeObject = storeObject
	committer.deduper.initialize(options, envRepo)
}

func (committer *committer) Commit(
	object *sku.Transacted,
	commitOptions sku.CommitOptions,
) (err error) {
	if err = committer.deduper.shouldCommit(object); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = committer.storeObject.Commit(
		object,
		commitOptions,
	); err != nil {
		err = errors.Wrap(err)
		return
	}
	return
}
