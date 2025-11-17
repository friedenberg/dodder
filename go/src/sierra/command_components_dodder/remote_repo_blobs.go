package command_components_dodder

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/repo_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/command"
	"code.linenisgreat.com/dodder/go/src/sierra/local_working_copy"
)

type RemoteRepoBlobs struct {
	EnvRepo
}

var _ interfaces.CommandComponentWriter = (*RemoteRepoBlobs)(nil)

func (cmd *RemoteRepoBlobs) SetFlagDefinitions(
	flagSet interfaces.CLIFlagDefinitions,
) {
}

// Returns a repo_blobs.BlobMutable that can be used to create a
// repo.Repo. The blob's public key SHOULD be set before writing it to the
// store.
func (cmd RemoteRepoBlobs) CreateRemoteBlob(
	req command.Request,
	local *local_working_copy.Repo,
	remoteType ids.Type,
) (blob repo_blobs.BlobMutable) {
	remoteEnvRepo := cmd.MakeEnvRepo(req, false)

	switch remoteType {
	default:
		errors.ContextCancelWithBadRequestf(
			req,
			"unsupported remote type: %q",
			remoteType,
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

	return blob
}
