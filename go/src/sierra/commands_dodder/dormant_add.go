package commands_dodder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/catgut"
	"code.linenisgreat.com/dodder/go/src/juliett/command"
	"code.linenisgreat.com/dodder/go/src/sierra/command_components_dodder"
)

func init() {
	utility.AddCmd("dormant-add", &DormantAdd{})
}

type DormantAdd struct {
	command_components_dodder.LocalWorkingCopy
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
