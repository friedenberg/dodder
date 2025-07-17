package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("dormant-add", &DormantAdd{})
}

type DormantAdd struct {
	command_components.LocalWorkingCopy
}

func (cmd DormantAdd) Run(dep command.Request) {
	localWorkingCopy := cmd.MakeLocalWorkingCopy(dep)

	localWorkingCopy.Must(errors.MakeFuncContextFromFuncErr(localWorkingCopy.Lock))

	for _, v := range dep.PopArgs() {
		cs := catgut.MakeFromString(v)

		if err := localWorkingCopy.GetDormantIndex().AddDormantTag(cs); err != nil {
			localWorkingCopy.Cancel(err)
		}
	}

	localWorkingCopy.Must(errors.MakeFuncContextFromFuncErr(localWorkingCopy.Unlock))
}
