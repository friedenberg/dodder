package command_components_dodder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/kilo/queries"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

type LocalWorkingCopyWithQueryGroup struct {
	LocalWorkingCopy
	Query
}

var _ interfaces.CommandComponentWriter = (*LocalWorkingCopyWithQueryGroup)(nil)

func (cmd *LocalWorkingCopyWithQueryGroup) SetFlagDefinitions(f interfaces.CLIFlagDefinitions) {
	cmd.LocalWorkingCopy.SetFlagDefinitions(f)
	cmd.Query.SetFlagDefinitions(f)
}

func (cmd LocalWorkingCopyWithQueryGroup) MakeLocalWorkingCopyAndQueryGroup(
	req command.Request,
	builderOptions queries.BuilderOption,
) (*local_working_copy.Repo, *queries.Query) {
	localWorkingCopy := cmd.MakeLocalWorkingCopy(req)

	queryGroup := cmd.MakeQueryIncludingWorkspace(
		req,
		builderOptions,
		localWorkingCopy,
		req.PopArgs(),
	)

	return localWorkingCopy, queryGroup
}
