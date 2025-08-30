package commands

import (
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
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

	for _, blobDigestString := range dep.PopArgs() {
		var blobDigest markl.Id

		if err := blobDigest.SetMaybeSha256(blobDigestString); err != nil {
			localWorkingCopy.Cancel(err)
		}

		objectIds, ok := lookupStored[string(blobDigest.GetBytes())]

		if ok {
			localWorkingCopy.GetUI().Printf(
				"%s (checked in as %q)",
				&blobDigest,
				objectIds,
			)
		} else {
			localWorkingCopy.GetUI().Printf("%s (missing)", &blobDigest)
		}
	}
}
