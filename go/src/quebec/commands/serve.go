package commands

import (
	"net"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/india/command_components_madder"
	"code.linenisgreat.com/dodder/go/src/oscar/remote_http"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
	"tailscale.com/client/local"
)

func init() {
	command.Register("serve", &Serve{})
}

type Serve struct {
	command_components.Env
	command_components_madder.EnvRepo
	command_components.LocalWorkingCopy

	TailscaleTLS bool
}

var _ interfaces.CommandComponentWriter = (*Serve)(nil)

func (cmd *Serve) SetFlagDefinitions(flagSet interfaces.CommandLineFlagDefinitions) {
	cmd.LocalWorkingCopy.SetFlagDefinitions(flagSet)

	flags.BoolVar(
		&cmd.TailscaleTLS,
		"tailscale-tls",
		false,
		"use tailscale for TLS",
	)
}

func (cmd Serve) Run(req command.Request) {
	args := req.PopArgs()
	errors.ContextSetCancelOnSIGHUP(req)

	envLocal := cmd.MakeEnvWithOptions(
		req,
		env_ui.Options{
			UIFileIsStderr: true,
			IgnoreTtyState: true,
		},
	)

	repo := cmd.MakeLocalWorkingCopyFromEnvLocal(envLocal)

	server := remote_http.Server{
		EnvLocal: envLocal,
		Repo:     repo,
	}

	if cmd.TailscaleTLS {
		var localClient local.Client
		server.GetCertificate = localClient.GetCertificate
	}

	// TODO switch network to be RemoteServeType
	var network, address string

	switch len(args) {
	case 0:
		network = "tcp"
		address = ":0"

	case 1:
		network = args[0]

	default:
		network = args[0]
		address = args[1]
	}

	if network == "-" {
		server.ServeStdio()
	} else {
		var listener net.Listener

		{
			var err error

			if listener, err = server.InitializeListener(
				network,
				address,
			); err != nil {
				envLocal.Cancel(err)
			}

			defer errors.ContextMustClose(envLocal, listener)
		}

		if err := server.Serve(listener); err != nil {
			envLocal.Cancel(err)
		}
	}
}
