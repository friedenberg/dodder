package commands

import (
	"flag"
	"strings"

	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
)

type Info struct{}

func init() {
	command.Register(
		"info",
		&Info{},
	)
}

func (c Info) SetFlagSet(f *flag.FlagSet) {}

func (c Info) Run(req command.Request) {
	dir := env_dir.MakeDefault(
		req,
		req.Debug,
	)

	ui := env_ui.Make(
		req,
		req.Config,
		env_ui.Options{},
	)

	args := req.PopArgs()

	if len(args) == 0 {
		args = []string{"store-version"}
	}

	defaultConfig := config_immutable.Default()

	for _, arg := range args {
		switch strings.ToLower(arg) {
		case "store-version":
			ui.GetUI().Print(defaultConfig.GetStoreVersion())

		case "store-version-next":
			ui.GetUI().Print(store_version.VNext)

		case "compression-type":
			ui.GetUI().Print(defaultConfig.GetBlobStoreConfigImmutable().GetBlobCompression())

		case "age-encryption":
			ui.GetUI().Print(defaultConfig.GetBlobStoreConfigImmutable().GetBlobEncryption())

		case "xdg":
			ecksDeeGee := dir.GetXDG()

			dotenv := xdg.Dotenv{
				XDG: &ecksDeeGee,
			}

			if _, err := dotenv.WriteTo(ui.GetOutFile()); err != nil {
				ui.CancelWithError(err)
			}
		}
	}
}
