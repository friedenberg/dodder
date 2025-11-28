package commands_dodder

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/kilo/command"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
	"code.linenisgreat.com/dodder/go/src/xray/command_components_dodder"
)

// TODO switch to using compound command pattern from blob_store_init.go
func init() {
	utility.AddCmd(
		"remote-add",
		&RemoteAdd{})
}

type RemoteAdd struct {
	command_components_dodder.LocalWorkingCopy
	command_components_dodder.RemoteTransfer

	complete command_components_dodder.Complete

	proto sku.Proto
}

var _ interfaces.CommandComponentWriter = (*RemoteAdd)(nil)

func (cmd *RemoteAdd) SetFlagDefinitions(flagSet interfaces.CLIFlagDefinitions) {
	cmd.RemoteTransfer.SetFlagDefinitions(flagSet)

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
	_, remoteObject := cmd.MakeRemoteAndObject(req, local)

	var id ids.RepoId

	if err := id.Set(req.PopArg("repo-id")); err != nil {
		req.Cancel(err)
	}

	req.AssertNoMoreArgs()

	if err := remoteObject.ObjectId.SetWithIdLike(&id); err != nil {
		req.Cancel(err)
	}

	// TODO connect to remote and get public key and validate

	cmd.proto.Apply(remoteObject.GetMetadataMutable(), genres.Repo)

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
