package commands

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("info-workspace", &InfoWorkspace{})
}

type InfoWorkspace struct {
	command_components.LocalWorkingCopy
}

func (cmd InfoWorkspace) Run(req command.Request) {
	repo := cmd.MakeLocalWorkingCopy(req)
	envWorkspace := repo.GetEnvWorkspace()
	envWorkspace.AssertNotTemporary(repo)
	args := req.PopArgs()

	if len(args) == 0 {
		args = []string{""}
	}

	for _, arg := range args {
		switch strings.ToLower(arg) {
		default:
			repo.CancelWithBadRequestf("unsupported info key: %q", arg)

		case "":
			// TODO what should this be?
			// TODO print toml representation?

		case "query":
			repo.GetUI().Print(
				envWorkspace.GetWorkspaceConfig().GetDefaultQueryGroup(),
			)

		case "defaults.type":
			repo.GetUI().Print(
				envWorkspace.GetWorkspaceConfig().GetDefaults().GetType(),
			)

		case "defaults.tags":
			repo.GetUI().Print(
				envWorkspace.GetWorkspaceConfig().GetDefaults().GetTags(),
			)
		}
	}
}
