package command_components_dodder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/cli"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/remote_connection_types"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/repo_blobs"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/oscar/remote_http"
)

type Remote struct {
	Env

	InventoryLists
	LocalWorkingCopy
	EnvRepo

	RemoteConnectionType remote_connection_types.Type
}

var _ interfaces.CommandComponentWriter = (*Remote)(nil)

func (cmd *Remote) SetFlagDefinitions(
	flagSet interfaces.CLIFlagDefinitions,
) {
	cli.FlagSetVarWithCompletion(
		flagSet,
		&cmd.RemoteConnectionType,
		// TODO rename to remote-connection-type?
		"remote-type",
	)
}

func (cmd Remote) CreateRemoteObject(
	req command.Request,
	local *local_working_copy.Repo,
) (remote repo.Repo, remoteObject *sku.Transacted) {
	remoteEnvRepo := cmd.MakeEnvRepo(req, false)
	remoteTypedRepoBlobStore := typed_blob_store.MakeRepoStore(remoteEnvRepo)

	remoteObject = sku.GetTransactedPool().Get()

	var blob repo_blobs.BlobMutable

	command.PopRequestArgTo(
		req.Args,
		"remote type",
		&remoteObject.Metadata.Type,
	)

	switch remoteObject.Metadata.Type {
	default:
		errors.ContextCancelWithBadRequestf(
			req,
			"unsupported remote type: %q",
			remoteObject.Metadata.Type,
		)

	case ids.GetOrPanic(ids.TypeTomlRepoLocalOverridePath).Type:
		xdgOverridePath := req.PopArg("xdg-path-override")

		blob = &repo_blobs.TomlLocalOverridePathV0{
			OverridePath: xdgOverridePath,
		}

	case ids.GetOrPanic(ids.TypeTomlRepoUri).Type:
		url := req.PopArg("url")

		var typedBlob repo_blobs.TomlUriV0

		if err := typedBlob.Uri.Set(url); err != nil {
			errors.ContextCancelWithBadRequestf(req, "invalid url: %s", err)
		}

		blob = &typedBlob

	case ids.GetOrPanic(ids.TypeTomlRepoLocalOverridePath).Type:
		path := req.PopArg("path")

		blob = &repo_blobs.TomlLocalOverridePathV0{
			OverridePath: remoteEnvRepo.AbsFromCwdOrSame(path),
		}
	}

	remote = cmd.MakeRemoteFromBlob(req, local, blob.GetRepoBlob())
	remoteConfig := remote.GetImmutableConfigPublic()
	blob.SetPublicKey(remoteConfig.GetPublicKey())

	var blobId interfaces.MarklId

	{
		var err error

		if blobId, _, err = remoteTypedRepoBlobStore.WriteTypedBlob(
			remoteObject.Metadata.Type,
			blob,
		); err != nil {
			req.Cancel(err)
		}
	}

	remoteObject.Metadata.GetBlobDigestMutable().ResetWithMarklId(blobId)

	return remote, remoteObject
}

func (cmd Remote) MakeRemote(
	req command.Request,
	repo *local_working_copy.Repo,
	object *sku.Transacted,
) (remote repo.Repo) {
	envRepo := cmd.MakeEnvRepo(req, false)
	typedRepoBlobStore := typed_blob_store.MakeRepoStore(envRepo)

	var blob repo_blobs.Blob

	{
		var err error

		if blob, _, err = typedRepoBlobStore.ReadTypedBlob(
			object.Metadata.Type,
			object.GetBlobDigest(),
		); err != nil {
			req.Cancel(err)
		}
	}

	remote = cmd.MakeRemoteFromBlob(req, repo, blob)

	return remote
}

func (cmd Remote) MakeRemoteFromBlob(
	req command.Request,
	repo *local_working_copy.Repo,
	blob repo_blobs.Blob,
) (remote repo.Repo) {
	env := cmd.MakeEnv(req)

	// TODO transform this to match how blob_store configs are turned into
	// objects
	// (by using interfaces instead of concrete types)
	switch blob := blob.(type) {
	case repo_blobs.BlobXDG:
		envDir := env_dir.MakeWithXDG(
			req,
			req.Config.Debug,
			blob.MakeXDG(req.Utility.GetName()),
		)

		envUI := env_ui.Make(
			req,
			req.Config,
			env.GetOptions(),
		)

		remote = local_working_copy.Make(
			env_local.Make(envUI, envDir),
			local_working_copy.OptionsEmpty,
		)

	case repo_blobs.TomlLocalOverridePathV0:
		envDir := env_dir.MakeWithXDGRootOverrideHomeAndInitialize(
			req,
			blob.OverridePath,
			req.Utility.GetName(),
			req.Config.Debug,
		)

		envUIOptions := env.GetOptions()
		envUIOptions.UIPrintingPrefix = "remote"

		envUI := env_ui.Make(
			req,
			req.Config,
			env.GetOptions(),
		)

		remote = local_working_copy.Make(
			env_local.Make(envUI, envDir),
			local_working_copy.OptionsEmpty,
		)
		// remote = cmd.MakeRemoteStdioLocal(
		// 	req,
		// 	env,
		// 	blob.OverridePath,
		// 	repo,
		// 	blob.GetPublicKey(),
		// )

	// case repo.RemoteTypeStdioSSH:
	// 	remote = cmd.MakeRemoteStdioSSH(
	// 		req,
	// 		env,
	// 		remoteArg,
	// 		local,
	// 	)

	// case repo.RemoteTypeSocketUnix:
	// 	remote = cmd.MakeRemoteHTTPFromXDGDotenvPath(
	// 		req,
	// 		remoteArg,
	// 		env.GetOptions(),
	// 		local,
	// 	)

	case repo_blobs.TomlUriV0:
		remote = cmd.MakeRemoteUrl(
			req,
			env,
			blob.Uri,
			repo,
		)

	default:
		errors.ContextCancelWithErrorf(req, "unsupported repo blob type: %T", blob)
	}

	return remote
}

func (cmd *Remote) MakeRemoteHTTPFromXDGDotenvPath(
	req command.Request,
	xdgDotenvPath string,
	options env_ui.Options,
	repo *local_working_copy.Repo,
	pubkey markl.Id,
) (remoteHTTP repo.Repo) {
	envLocal := cmd.MakeEnvWithXDGLayoutAndOptions(
		req,
		xdgDotenvPath,
		options,
	)

	remote := cmd.MakeLocalWorkingCopyFromEnvLocal(envLocal)

	server := &remote_http.Server{
		EnvLocal: envLocal,
		Repo:     remote,
	}

	var httpRoundTripper remote_http.RoundTripperUnixSocket

	if err := httpRoundTripper.Initialize(
		server,
		pubkey,
	); err != nil {
		req.Cancel(err)
	}

	go func() {
		if err := server.Serve(httpRoundTripper.UnixSocket); err != nil {
			req.Cancel(err)
		}
	}()

	remoteHTTP = remote_http.MakeClient(
		envLocal,
		&httpRoundTripper,
		repo,
		cmd.MakeInventoryListCoderCloset(repo.GetEnvRepo()),
	)

	return remoteHTTP
}

func (cmd *Remote) MakeRemoteStdioSSH(
	req command.Request,
	env env_local.Env,
	arg string,
	repo *local_working_copy.Repo,
) (remoteHTTP repo.Repo) {
	envRepo := cmd.MakeEnvRepo(req, false)

	var httpRoundTripper remote_http.RoundTripperStdio

	if err := httpRoundTripper.InitializeWithSSH(
		envRepo,
		arg,
	); err != nil {
		env.Cancel(err)
	}

	remoteHTTP = remote_http.MakeClient(
		envRepo,
		&httpRoundTripper,
		repo,
		cmd.MakeInventoryListCoderCloset(envRepo),
	)

	return remoteHTTP
}

func (cmd *Remote) MakeRemoteStdioLocal(
	req command.Request,
	env env_local.Env,
	dir string,
	repo *local_working_copy.Repo,
	pubkey interfaces.MarklId,
) (remoteHTTP repo.Repo) {
	envRepo := cmd.MakeEnvRepo(req, false)

	var httpRoundTripper remote_http.RoundTripperStdio

	if err := files.AssertDir(dir); err != nil {
		if files.IsErrNotDirectory(err) {
			errors.ContextCancelWithBadRequestError(req, err)
		} else {
			req.Cancel(err)
		}
	}

	httpRoundTripper.Cmd.Dir = dir

	if err := httpRoundTripper.InitializeWithLocal(
		envRepo,
		repo.GetConfig(),
		pubkey,
	); err != nil {
		env.Cancel(err)
	}

	remoteHTTP = remote_http.MakeClient(
		env,
		&httpRoundTripper,
		repo,
		cmd.MakeInventoryListCoderCloset(envRepo),
	)

	return remoteHTTP
}

func (cmd *Remote) MakeRemoteUrl(
	req command.Request,
	env env_local.Env,
	uri values.Uri,
	repo *local_working_copy.Repo,
) (remoteHTTP repo.Repo) {
	envRepo := cmd.MakeEnvRepo(req, false)

	remoteHTTP = remote_http.MakeClient(
		envRepo,
		&remote_http.RoundTripperHost{
			UrlData:      remote_http.MakeUrlDataFromUri(uri),
			RoundTripper: remote_http.DefaultRoundTripper,
		},
		repo,
		cmd.MakeInventoryListCoderCloset(envRepo),
	)

	return remoteHTTP
}
