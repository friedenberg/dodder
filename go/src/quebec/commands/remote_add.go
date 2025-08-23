package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

// TODO switch to using compound command pattern from blob_store_init.go
func init() {
	command.Register(
		"remote-add",
		&RemoteAdd{},
	)
}

type RemoteAdd struct {
	command_components.LocalWorkingCopy
	command_components.RemoteTransfer

	complete command_components.Complete

	proto sku.Proto
}

func (cmd *RemoteAdd) SetFlagSet(flagSet *flags.FlagSet) {
	cmd.RemoteTransfer.SetFlagSet(flagSet)

	flagSet.Var(
		cmd.complete.GetFlagValueMetadataTags(&cmd.proto.Metadata),
		"tags",
		"tags added for new objects in `checkin`, `new`, `organize`",
	)

	cmd.proto.SetFlagSetDescription(
		flagSet,
		"description to use for the new repo",
	)
}

func (cmd RemoteAdd) Run(req command.Request) {
	local := cmd.MakeLocalWorkingCopy(req)
	_, remoteObject := cmd.CreateRemoteObject(req, local)

	var id ids.RepoId

	if err := id.Set(req.PopArg("repo-id")); err != nil {
		req.Cancel(err)
	}

	req.AssertNoMoreArgs()

	if err := remoteObject.ObjectId.SetWithIdLike(&id); err != nil {
		req.Cancel(err)
	}

	// TODO connect to remote and get public key and validate

	cmd.proto.Apply(remoteObject.GetMetadata(), genres.Repo)

	req.Must(errors.MakeFuncContextFromFuncErr(local.Lock))

	if err := local.GetStore().CreateOrUpdateDefaultProto(
		remoteObject,
		sku.StoreOptions{
			ApplyProto: true,
		},
	); err != nil {
		req.Cancel(err)
	}

	req.Must(errors.MakeFuncContextFromFuncErr(local.Unlock))
}
