package commands

import (
	"bufio"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/env_vars"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
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

func (cmd Info) SetFlagSet(flagSet interfaces.CommandLineFlagDefinitions) {}

func (cmd Info) Run(req command.Request) {
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

	defaultGenesisConfig := genesis_configs.DefaultWithVersion(
		store_version.VCurrent,
		ids.TypeInventoryListVCurrent,
	).Blob

	defaultBlobStoreConfig := blob_store_configs.Default().Blob

	for _, arg := range args {
		// TODO switch to underscore+hyphen string keys
		switch strings.ToLower(arg) {
		case "store-version":
			ui.GetUI().Print(defaultGenesisConfig.GetStoreVersion())

		case "store-version-next":
			ui.GetUI().Print(store_version.VNext)

		case "compression-type":
			if ioWrapper, ok := defaultBlobStoreConfig.(interfaces.BlobIOWrapper); ok {
				ui.GetUI().Print(
					ioWrapper.GetBlobCompression(),
				)
			} else {
				errors.ContextCancelWithBadRequestf(ui, "default blob store does not support compression")
			}

		case "age-encryption":
			if ioWrapper, ok := defaultBlobStoreConfig.(interfaces.BlobIOWrapper); ok {
				ui.GetUI().Print(
					ioWrapper.GetBlobEncryption(),
				)
			} else {
				errors.ContextCancelWithBadRequestf(ui, "default blob store does not support encryption")
			}

		case "env":
			envVars := env_vars.Make(dir)
			var coder env_vars.BufferedCoderDotenv
			bufferedWriter := bufio.NewWriter(ui.GetOutFile())

			if _, err := coder.EncodeTo(envVars, bufferedWriter); err != nil {
				ui.Cancel(err)
			}

			if err := bufferedWriter.Flush(); err != nil {
				ui.Cancel(err)
			}

		case "xdg":
			ecksDeeGee := dir.GetXDG()
			envVars := env_vars.Make(ecksDeeGee)
			var coder env_vars.BufferedCoderDotenv
			bufferedWriter := bufio.NewWriter(ui.GetOutFile())

			if _, err := coder.EncodeTo(envVars, bufferedWriter); err != nil {
				ui.Cancel(err)
			}

			if err := bufferedWriter.Flush(); err != nil {
				ui.Cancel(err)
			}
		}
	}
}
