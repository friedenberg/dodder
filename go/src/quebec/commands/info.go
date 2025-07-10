package commands

import (
	"bufio"
	"flag"
	"strings"

	"code.linenisgreat.com/dodder/go/src/bravo/env_vars"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_config"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_config"
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

	defaultGenesisConfig := genesis_config.Default()
	defaultBlobStoreConfig := blob_store_config.Default()

	for _, arg := range args {
		// TODO switch to underscore+hyphen string keys
		switch strings.ToLower(arg) {
		case "store-version":
			ui.GetUI().Print(defaultGenesisConfig.GetStoreVersion())

		case "store-version-next":
			ui.GetUI().Print(store_version.VNext)

		case "compression-type":
			ui.GetUI().Print(
				defaultBlobStoreConfig.GetBlobCompression(),
			)

		case "age-encryption":
			ui.GetUI().Print(
				defaultBlobStoreConfig.GetBlobEncryption(),
			)

		case "env":
			envVars := env_vars.Make(dir)
			var coder env_vars.BufferedCoderDotenv
			bufferedWriter := bufio.NewWriter(ui.GetOutFile())

			if _, err := coder.EncodeTo(envVars, bufferedWriter); err != nil {
				ui.CancelWithError(err)
			}

			if err := bufferedWriter.Flush(); err != nil {
				ui.CancelWithError(err)
			}

		case "xdg":
			ecksDeeGee := dir.GetXDG()
			envVars := env_vars.Make(ecksDeeGee)
			var coder env_vars.BufferedCoderDotenv
			bufferedWriter := bufio.NewWriter(ui.GetOutFile())

			if _, err := coder.EncodeTo(envVars, bufferedWriter); err != nil {
				ui.CancelWithError(err)
			}

			if err := bufferedWriter.Flush(); err != nil {
				ui.CancelWithError(err)
			}
		}
	}
}
