package commands

import (
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("find-missing", &FindMissing{})
}

type FindMissing struct {
	command_components.LocalWorkingCopy
}

func (cmd FindMissing) Run(dep command.Request) {
	localWorkingCopy := cmd.MakeLocalWorkingCopy(dep)

	var lookupStored map[string][]string

	{
		var err error

		if lookupStored, err = localWorkingCopy.GetStore().MakeBlobDigestBytesMap(); err != nil {
			dep.Cancel(err)
		}
	}

	for _, digestString := range dep.PopArgs() {
		var digest sha.Sha

		if err := digest.Set(digestString); err != nil {
			localWorkingCopy.Cancel(err)
		}

		objectIds, ok := lookupStored[string(digest.GetBytes())]

		if ok {
			localWorkingCopy.GetUI().Printf(
				"%s (checked in as %q)",
				&digest,
				objectIds,
			)
		} else {
			localWorkingCopy.GetUI().Printf("%s (missing)", &digest)
		}
	}
}
