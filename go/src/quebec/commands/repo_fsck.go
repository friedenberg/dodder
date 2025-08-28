package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register(
		"repo-fsck",
		&RepoFsck{},
	)
}

type RepoFsck struct {
	command_components.LocalWorkingCopy
	command_components.EnvRepo
	command_components.BlobStore
}

// TODO add completion for blob store id's

func (cmd RepoFsck) Run(req command.Request) {
	req.AssertNoMoreArgs()

	repo := cmd.MakeLocalWorkingCopyWithOptions(
		req,
		env_ui.Options{},
		local_working_copy.OptionsAllowConfigReadError,
	)

	store := repo.GetStore()
	missingObjects := sku.MakeListTransacted()

	for objectWithList, err := range store.GetInventoryListStore().IterAllSkus() {
		errors.ContextContinueOrPanic(repo)

		if err == nil {
			continue
		}

		if env_dir.IsErrBlobMissing(err) {
			missingObjects.Add(objectWithList.List)
			continue
		}

		unwrapped := errors.Unwrap(err)

		if unwrapped != nil {
			repo.GetErr().Print(unwrapped)
		} else {
			err = errors.Wrapf(
				err,
				"List: %s",
				sku.String(objectWithList.List),
			)

			ui.CLIErrorTreeEncoder.EncodeTo(err, repo.GetErr())
		}

	}

	repo.GetUI().Print("missing list blobs: ")

	for missingList := range missingObjects.All() {
		repo.GetUI().Print(sku.String(missingList))
	}
}
