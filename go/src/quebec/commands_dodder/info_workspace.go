package commands_dodder

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/workspace_config_blobs"
	"code.linenisgreat.com/dodder/go/src/papa/command_components_dodder"
)

func init() {
	utility.AddCmd("info-workspace", &InfoWorkspace{})
}

// TODO rename to WorkspaceInfo
type InfoWorkspace struct {
	command_components_dodder.LocalWorkingCopy
}

func (cmd InfoWorkspace) Run(req command.Request) {
	repo := cmd.MakeLocalWorkingCopy(req)
	envWorkspace := repo.GetEnvWorkspace()
	envWorkspace.AssertNotTemporary(repo)
	arg := req.PopArgOrDefault("workspace info key", "")
	req.AssertNoMoreArgs()

	// TODO convert arg to flag.Value type that supports completion
	switch strings.ToLower(arg) {
	default:
		errors.ContextCancelWithBadRequestf(
			repo,
			"unsupported info key: %q",
			arg,
		)

	case "":
		// TODO what should this be?
		// TODO print toml representation?

	case "query":
		workspaceConfig := envWorkspace.GetWorkspaceConfig()

		type WithQueryGroup = workspace_config_blobs.ConfigWithDefaultQueryString

		if withQueryGroup, ok := workspaceConfig.(WithQueryGroup); ok {
			repo.GetUI().Print(
				withQueryGroup.GetDefaultQueryString(),
			)
		} else {
			errors.ContextCancelWithBadRequestf(repo, "workspace does not support default queries")
		}

	case "defaults.type":
		repo.GetUI().Print(
			envWorkspace.GetWorkspaceConfig().GetDefaults().GetDefaultType(),
		)

	case "defaults.tags":
		repo.GetUI().Print(
			envWorkspace.GetWorkspaceConfig().GetDefaults().GetDefaultTags(),
		)
	}
}
