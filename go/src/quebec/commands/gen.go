package commands

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
)

type Gen struct{}

func init() {
	command.Register(
		"gen",
		&Gen{},
	)
}

func (cmd Gen) SetFlagSet(flagSet interfaces.CommandLineFlagDefinitions) {}

func (cmd Gen) Run(req command.Request) {
	envUI := env_ui.Make(
		req,
		req.Config,
		env_ui.Options{},
	)

	args := req.PopArgs()

	for _, arg := range args {
		arg = strings.ToLower(arg)

		switch arg {
		case markl.PurposeMadderPrivateKeyV0:
			var id markl.Id

			if err := markl.GeneratePrivateKey(
				nil,
				arg,
				markl.FormatIdAgeX25519Sec,
				&id,
			); err != nil {
				ui.Err().Print(err)
				continue
			}

			envUI.GetUI().Print(id.StringWithFormat())

		case markl.PurposeMadderPrivateKeyV1:
			var id markl.Id

			if err := markl.GeneratePrivateKey(
				nil,
				arg,
				markl.FormatIdAgeX25519Sec,
				&id,
			); err != nil {
				ui.Err().Print(err)
				continue
			}

			envUI.GetUI().Print(id.StringWithFormat())

		case markl.PurposeRepoPrivateKeyV1:
			var id markl.Id

			if err := markl.GeneratePrivateKey(
				nil,
				arg,
				markl.FormatIdEd25519Sec,
				&id,
			); err != nil {
				ui.Err().Print(err)
				continue
			}

			envUI.GetUI().Print(id.StringWithFormat())
		}
	}
}
