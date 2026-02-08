package commands_dodder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_dir"
	"code.linenisgreat.com/dodder/go/src/juliett/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/command_components_madder"
	"code.linenisgreat.com/dodder/go/src/victor/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/xray/command_components_dodder"
)

func init() {
	utility.AddCmd(
		"repo-fsck",
		&RepoFsck{})
}

type RepoFsck struct {
	command_components_dodder.LocalWorkingCopy
	command_components_dodder.EnvRepo
	command_components_madder.BlobStore
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

	for objectWithList, err := range store.GetInventoryListStore().AllInventoryListObjectsAndContents() {
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
