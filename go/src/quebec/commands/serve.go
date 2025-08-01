package commands

import (
	"flag"
	"net"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/oscar/remote_http"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
	"tailscale.com/client/local"
)

func init() {
	command.Register("serve", &Serve{})
}

type Serve struct {
	command_components.Env
	command_components.EnvRepo
	command_components.LocalWorkingCopy

	TailscaleTLS bool
}

func (cmd *Serve) SetFlagSet(flagSet *flag.FlagSet) {
	cmd.LocalWorkingCopy.SetFlagSet(flagSet)

	flag.BoolVar(
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
		var lc local.Client
		server.GetCertificate = lc.GetCertificate
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
