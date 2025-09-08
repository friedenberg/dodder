package blob_stores

import (
	"fmt"
	"net"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// TODO refactor `blob_store_configs.ConfigSFTP` for ssh-client-specific methods
func MakeSSHClientForExplicitConfig(
	ctx interfaces.ActiveContext,
	uiPrinter ui.Printer,
	config blob_store_configs.ConfigSFTPConfigExplicit,
) (sshClient *ssh.Client, err error) {
	clientConfig := &ssh.ClientConfig{
		User:            config.GetUser(),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: make this configurable
	}

	// Configure authentication
	if config.GetPrivateKeyPath() != "" {
		var key ssh.Signer
		var keyBytes []byte

		if keyBytes, err = os.ReadFile(config.GetPrivateKeyPath()); err != nil {
			err = errors.Wrapf(err, "failed to read private key")
			return
		}

		if key, err = ssh.ParsePrivateKey(keyBytes); err != nil {
			err = errors.Wrapf(err, "failed to parse private key")
			return
		}

		clientConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(key)}
	} else if config.GetPassword() != "" {
		clientConfig.Auth = []ssh.AuthMethod{ssh.Password(config.GetPassword())}
	} else {
		err = errors.Errorf("no authentication method configured")
		return
	}

	addr := fmt.Sprintf(
		"%s:%d",
		config.GetHost(),
		config.GetPort(),
	)

	if sshClient, err = sshDial(ctx, uiPrinter, clientConfig, addr); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func MakeSSHClientFromSSHConfig(
	ctx interfaces.ActiveContext,
	uiPrinter ui.Printer,
	config blob_store_configs.ConfigSFTPUri,
) (sshClient *ssh.Client, err error) {
	socket := os.Getenv("SSH_AUTH_SOCK")

	if socket == "" {
		err = errors.Errorf("SSH_AUTH_SOCK empty or unset")
		return
	}

	var connSshSock net.Conn

	ui.Log().Print("connecting to SSH_AUTH_SOCK: %s", socket)
	if connSshSock, err = net.Dial("unix", socket); err != nil {
		err = errors.Wrapf(err, "failed to connect to SSH_AUTH_SOCK")
		return
	}

	ctx.After(errors.MakeFuncContextFromFuncErr(connSshSock.Close))

	ui.Log().Print("creating ssh-agent client")
	clientAgent := agent.NewClient(connSshSock)

	uri := config.GetUri()
	url := uri.GetUrl()

	clientConfig := &ssh.ClientConfig{
		User: url.User.Username(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(clientAgent.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO make configurable
	}

	addr := fmt.Sprintf(
		"%s:%s",
		url.Hostname(),
		url.Port(),
	)

	if sshClient, err = sshDial(ctx, uiPrinter, clientConfig, addr); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func sshDial(
	ctx interfaces.ActiveContext,
	uiPrinter ui.Printer,
	configClient *ssh.ClientConfig,
	addr string,
) (sshClient *ssh.Client, err error) {
	uiPrinter.Printf("dialing %q...", addr)
	if sshClient, err = ssh.Dial("tcp", addr, configClient); err != nil {
		err = errors.Wrapf(err, "failed to connect to SSH server")
		return
	}
	uiPrinter.Printf("connected to %q...", addr)

	ctx.After(errors.MakeFuncContextFromFuncErr(sshClient.Close))

	return
}
