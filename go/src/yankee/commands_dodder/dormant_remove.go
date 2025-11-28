package commands_dodder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/catgut"
	"code.linenisgreat.com/dodder/go/src/kilo/command"
	"code.linenisgreat.com/dodder/go/src/xray/command_components_dodder"
)

func init() {
	utility.AddCmd("dormant-remove", &DormantRemove{})
}

type DormantRemove struct {
	command_components_dodder.LocalWorkingCopy
}

func (cmd DormantRemove) Run(dep command.Request) {
	localWorkingCopy := cmd.MakeLocalWorkingCopy(dep)
	localWorkingCopy.Must(errors.MakeFuncContextFromFuncErr(localWorkingCopy.Lock))

	for _, v := range dep.PopArgs() {
		cs := catgut.MakeFromString(v)

		if err := localWorkingCopy.GetDormantIndex().RemoveDormantTag(
			cs,
		); err != nil {
			localWorkingCopy.Cancel(err)
		}
	}

	localWorkingCopy.Must(errors.MakeFuncContextFromFuncErr(localWorkingCopy.Unlock))
}
