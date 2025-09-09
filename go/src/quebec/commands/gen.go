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
		case markl.FormatIdMadderPrivateKeyV0:
			var id markl.Id

			if err := markl.GeneratePrivateKey(
				nil,
				arg,
				markl.TypeIdAgeX25519Sec,
				&id,
			); err != nil {
				ui.Err().Print(err)
				continue
			}

			envUI.GetUI().Print(id.StringWithFormat())

		case markl.FormatIdMadderPrivateKeyV1:
			var id markl.Id

			if err := markl.GeneratePrivateKey(
				nil,
				arg,
				markl.TypeIdAgeX25519Sec,
				&id,
			); err != nil {
				ui.Err().Print(err)
				continue
			}

			envUI.GetUI().Print(id.StringWithFormat())

		case markl.FormatIdRepoPrivateKeyV1:
			var id markl.Id

			if err := markl.GeneratePrivateKey(
				nil,
				arg,
				markl.TypeIdEd25519Sec,
				&id,
			); err != nil {
				ui.Err().Print(err)
				continue
			}

			envUI.GetUI().Print(id.StringWithFormat())
		}
	}
}
