package remote_transfer

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
)

type committer struct {
	options     repo.ImporterOptions
	storeObject sku.RepoStore
	deduper     deduper
}

func (committer *committer) initialize(
	options repo.ImporterOptions,
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
