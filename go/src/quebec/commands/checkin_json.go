package commands

import (
	"encoding/json"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("checkin-json", &CheckinJson{})
}

type CheckinJson struct {
	command_components.LocalWorkingCopy
}

type TomlBookmark struct {
	ObjectId string
	Tags     []string
	Url      string
}

func (cmd CheckinJson) Run(req command.Request) {
	localWorkingCopy := cmd.MakeLocalWorkingCopy(req)

	dec := json.NewDecoder(localWorkingCopy.GetInFile())

	for {
		var entry TomlBookmark

		if err := dec.Decode(&entry); err != nil {
			if errors.IsEOF(err) {
				err = nil
				break
			} else {
				localWorkingCopy.Cancel(err)
			}
		}
	}
}
