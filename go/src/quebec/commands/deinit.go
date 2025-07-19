package commands

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"code.linenisgreat.com/dodder/go/src/bravo/comments"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("deinit", &Deinit{})
}

type Deinit struct {
	command_components.LocalArchive

	Force bool
}

func (cmd *Deinit) SetFlagSet(f *flag.FlagSet) {
	f.BoolVar(
		&cmd.Force,
		"force",
		false,
		"force deinit",
	)
}

func (cmd Deinit) Run(req command.Request) {
	envRepo := cmd.MakeEnvRepo(req, false)
	localWorkingCopy := cmd.MakeLocalArchive(envRepo)

	var home string

	{
		var err error

		if home, err = os.UserHomeDir(); err != nil {
			req.Cancel(err)
		}

		if home, err = filepath.Abs(home); err != nil {
			req.Cancel(err)
		}
	}

	exdg := envRepo.GetXDG()

	comments.Comment(
		"determine if this is a native XDG repo, or an overridden XDG repo",
	)

	var xdgHome string

	{
		var err error

		if xdgHome, err = filepath.Abs(exdg.Home); err != nil {
			req.Cancel(err)
		}
	}

	var filesAndDirectories []string

	if xdgHome == home {
		filesAndDirectories = localWorkingCopy.GetEnvRepo().GetXDG().GetXDGPaths()
	} else {
		filesAndDirectories = []string{xdgHome}
	}

	sort.Strings(filesAndDirectories)

	if repo, ok := localWorkingCopy.(repo.LocalWorkingCopy); ok &&
		!repo.GetEnvWorkspace().IsTemporary() {
		workspaceConfigFilePath := repo.GetEnvWorkspace().GetWorkspaceConfigFilePath()

		if workspaceConfigFilePath != "" {
			filesAndDirectories = append(
				filesAndDirectories,
				workspaceConfigFilePath,
			)
		}
	}
	
	// TODO decide whether the workspace directory should be deleted too

	if !cmd.Force &&
		!envRepo.Confirm(
			fmt.Sprintf(
				`are you sure you want to deinit?
The following directories and files would be deleted:

%s`, strings.Join(filesAndDirectories, "\n")),
		) {
		ui.Err().Print("permission denied and -force not specified, aborting")
		return
	}

	base := path.Join(localWorkingCopy.GetEnvRepo().Dir())

	if err := files.SetAllowUserChangesRecursive(base); err != nil {
		envRepo.Cancel(err)
	}

	if err := envRepo.Delete(
		filesAndDirectories...,
	); err != nil {
		envRepo.Cancel(err)
	}
}
