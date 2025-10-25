package commands_dodder

import (
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/hotel/workspace_config_blobs"
	"code.linenisgreat.com/dodder/go/src/india/command_components"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/papa/command_components_dodder"
)

func init() {
	utility.AddCmd(
		"init-workspace",
		&InitWorkspace{})
}

type InitWorkspace struct {
	command_components.Env
	command_components_dodder.LocalWorkingCopy

	complete command_components_dodder.Complete

	DefaultQueryGroup values.String
	Proto             sku.Proto
}

var _ interfaces.CommandComponentWriter = (*InitWorkspace)(nil)

func (cmd *InitWorkspace) SetFlagDefinitions(
	flagSet interfaces.CLIFlagDefinitions,
) {
	cmd.LocalWorkingCopy.SetFlagDefinitions(flagSet)
	// TODO add command.Completer variants of tags, type, and query flags

	flagSet.Var(
		cmd.complete.GetFlagValueMetadataTags(&cmd.Proto.Metadata),
		"tags",
		"tags added for new objects in `checkin`, `new`, `organize`",
	)

	flagSet.Var(
		cmd.complete.GetFlagValueMetadataType(&cmd.Proto.Metadata),
		"type",
		"type used for new objects in `new` and `organize`",
	)

	flagSet.Var(
		cmd.complete.GetFlagValueStringTags(&cmd.DefaultQueryGroup),
		"query",
		"default query for `show`",
	)
}

func (cmd InitWorkspace) Complete(
	_ command.Request,
	envLocal env_local.Env,
	commandLine command.CommandLine,
) {
	searchDir := envLocal.GetCwd()

	if commandLine.InProgress != "" && files.Exists(commandLine.InProgress) {
		var err error

		if commandLine.InProgress, err = filepath.Abs(commandLine.InProgress); err != nil {
			envLocal.Cancel(err)
			return
		}

		if searchDir, err = filepath.Rel(searchDir, commandLine.InProgress); err != nil {
			envLocal.Cancel(err)
			return
		}
	}

	for dirEntry, err := range files.WalkDir(searchDir) {
		if err != nil {
			envLocal.Cancel(err)
			return
		}

		if !dirEntry.IsDir() {
			continue
		}

		if files.WalkDirIgnoreFuncHidden(dirEntry) {
			continue
		}

		envLocal.GetUI().Printf("%s/\tdirectory", dirEntry.RelPath)
	}
}

func (cmd InitWorkspace) Run(req command.Request) {
	envLocal := cmd.MakeEnv(req)

	switch req.RemainingArgCount() {
	case 0:
		break

	case 1:
		dir := req.PopArg("dir")

		if err := envLocal.MakeDirs(dir); err != nil {
			req.Cancel(err)
			return
		}

		if err := os.Chdir(dir); err != nil {
			req.Cancel(err)
			return
		}
	}

	req.AssertNoMoreArgs()

	localWorkingCopy := cmd.MakeLocalWorkingCopy(req)

	blob := &workspace_config_blobs.V0{
		Query: cmd.DefaultQueryGroup.String(),
		Defaults: repo_configs.DefaultsV1OmitEmpty{
			Type: cmd.Proto.Type,
			Tags: quiter.Elements(cmd.Proto.Tags),
		},
	}

	if err := localWorkingCopy.GetEnvWorkspace().CreateWorkspace(
		blob,
	); err != nil {
		req.Cancel(err)
	}
}
