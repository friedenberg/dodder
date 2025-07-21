package blob_stores

import (
	"fmt"
	"io"
	"net"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var defaultBuckets = []int{2}

// TODO describe base path agnostically
func MakeBlobStore(
	ctx interfaces.Context,
	basePath string,
	config blob_store_configs.Config,
	tempFS env_dir.TemporaryFS,
) (store interfaces.LocalBlobStore, err error) {
	switch tipe := config.GetBlobStoreType(); tipe {
	default:
		err = errors.BadRequestf("unsupported blob store type %q", tipe)
		return

	case "sftp":
		var sshClient *ssh.Client
		var configSFTP blob_store_configs.ConfigSFTPRemotePath

		switch config := config.(type) {
		default:
			err = errors.BadRequestf("unsupported blob store config for type %q: %T", tipe, config)
			return

		case blob_store_configs.ConfigSFTPUri:
			if sshClient, err = MakeSSHClientFromSSHConfig(ctx, config); err != nil {
				err = errors.Wrap(err)
				return
			}

			configSFTP = config

		case blob_store_configs.ConfigSFTPConfigExplicit:
			if sshClient, err = MakeSSHClientForExplicitConfig(ctx, config); err != nil {
				err = errors.Wrap(err)
				return
			}

			configSFTP = config
		}

		return makeSftpStore(ctx, configSFTP, sshClient)

	case "local":
		if config, ok := config.(blob_store_configs.ConfigLocalHashBucketed); ok {
			return makeLocalHashBucketed(ctx, basePath, config, tempFS)
		} else {
			err = errors.BadRequestf("unsupported blob store config for type %q: %T", tipe, config)
			return
		}
	}
}

func CopyBlobIfNecessary(
	env env_ui.Env,
	dst interfaces.BlobStore,
	src interfaces.BlobStore,
	blobShaGetter interfaces.Digester,
	extraWriter io.Writer,
) (n int64, err error) {
	if src == nil {
		return
	}

	blobSha := blobShaGetter.GetDigest()

	if dst.HasBlob(blobSha) || blobSha.IsNull() {
		err = env_dir.MakeErrAlreadyExists(
			blobSha,
			"",
		)

		return
	}

	return CopyBlob(env, dst, src, blobSha, extraWriter)
}

// TODO make this honor context closure and abort early
func CopyBlob(
	env env_ui.Env,
	dst interfaces.BlobStore,
	src interfaces.BlobStore,
	blobSha interfaces.Digest,
	extraWriter io.Writer,
) (n int64, err error) {
	if src == nil {
		return
	}

	var rc interfaces.ReadCloseDigester

	if rc, err = src.BlobReader(blobSha); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.ContextMustClose(env, rc)

	var wc interfaces.WriteCloseDigester

	if wc, err = dst.BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO should this be closed with an error when the shas don't match to
	// prevent a garbage object in the store?
	defer errors.ContextMustClose(env, wc)

	outputWriter := io.Writer(wc)

	if extraWriter != nil {
		outputWriter = io.MultiWriter(outputWriter, extraWriter)
	}

	if n, err = io.Copy(outputWriter, rc); err != nil {
		err = errors.Wrap(err)
		return
	}

	shaRc := rc.GetDigest()
	shaWc := wc.GetDigest()

	if !digests.DigestEquals(shaRc, blobSha) ||
		!digests.DigestEquals(shaWc, blobSha) {
		err = errors.ErrorWithStackf(
			"lookup sha was %s, read sha was %s, but written sha was %s",
			blobSha,
			shaRc,
			shaWc,
		)
	}

	return
}

// TODO offer options like just checking the existence of the blob, getting its
// size, or full verification
func VerifyBlob(
	ctx interfaces.Context,
	blobStore interfaces.LocalBlobStore,
	sh interfaces.Digest,
	progressWriter io.Writer,
) (err error) {
	// TODO check if `blobStore` implements a `VerifyBlob` method and call that
	// instead (for expensive blob stores that may implement their own remote
	// verification, such as ssh, sftp, or something else)

	var readCloser interfaces.ReadCloseDigester

	if readCloser, err = blobStore.BlobReader(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	if _, err = io.Copy(progressWriter, readCloser); err != nil {
		err = errors.Wrap(err)
		return
	}

	expected := sha.MustWithDigest(sh)

	if err = expected.AssertEqualsShaLike(readCloser.GetDigest()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = readCloser.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO refactor `blob_store_configs.ConfigSFTP` for ssh-client-specific methods
func MakeSSHClientForExplicitConfig(
	ctx interfaces.Context,
	config blob_store_configs.ConfigSFTPConfigExplicit,
) (sshClient *ssh.Client, err error) {
	sshConfig := &ssh.ClientConfig{
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

		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(key)}
	} else if config.GetPassword() != "" {
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(config.GetPassword())}
	} else {
		err = errors.Errorf("no authentication method configured")
		return
	}

	addr := fmt.Sprintf(
		"%s:%d",
		config.GetHost(),
		config.GetPort(),
	)

	if sshClient, err = ssh.Dial("tcp", addr, sshConfig); err != nil {
		err = errors.Wrapf(err, "failed to connect to SSH server")
		return
	}

	ctx.After(errors.MakeFuncContextFromFuncErr(sshClient.Close))

	return
}

func MakeSSHClientFromSSHConfig(
	ctx interfaces.Context,
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

	configClient := &ssh.ClientConfig{
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

	ui.Log().Printf("connecting via ssh: %q", addr)
	if sshClient, err = ssh.Dial("tcp", addr, configClient); err != nil {
		err = errors.Wrapf(err, "failed to connect to SSH server")
		return
	}

	ctx.After(errors.MakeFuncContextFromFuncErr(sshClient.Close))

	return
}
