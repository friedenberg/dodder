package commands

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_ptr"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register(
		"checkin-blob",
		&CheckinBlob{
			NewTags: collections_ptr.MakeFlagCommas[ids.Tag](
				collections_ptr.SetterPolicyAppend,
			),
		},
	)
}

type CheckinBlob struct {
	command_components.LocalWorkingCopy

	Delete  bool
	NewTags collections_ptr.Flag[ids.Tag, *ids.Tag]
}

func (cmd *CheckinBlob) SetFlagSet(f *flag.FlagSet) {
	f.BoolVar(&cmd.Delete, "delete", false, "the checked-out file")
	f.Var(
		cmd.NewTags,
		"new-tags",
		"comma-separated tags (will replace existing tags)",
	)
}

func (cmd CheckinBlob) Run(req command.Request) {
	args := req.PopArgs()

	localWorkingCopy := cmd.MakeLocalWorkingCopy(req)

	if len(args)%2 != 0 {
		errors.ContextCancelWithErrorf(
			req,
			"arguments must come in pairs of zettel id and blob path or sha",
		)
	}

	pairs := make([]externalBlobPair, len(args)/2)

	// transform args into pairs of object id's and filepaths or shas
	for idx, pair := range pairs {
		// TODO switch to using object ID instead to allow
		zettelIdString := args[idx*2]
		filepathOrSha := args[(idx*2)+1]

		if err := pair.SetArgs(
			zettelIdString,
			filepathOrSha,
			localWorkingCopy.GetEnvRepo(),
		); err != nil {
			req.Cancel(err)
		}

		pairs[idx] = pair
	}

	for idx, pair := range pairs {
		// iterate through pairs and read current zettel
		{
			var err error

			if pairs[idx].object, err = localWorkingCopy.GetStore().ReadTransactedFromObjectId(
				pair.ZettelId,
			); err != nil {
				req.Cancel(err)
			}
		}

		object := pairs[idx].object

		if err := object.SetBlobSha(pair.GetDigest()); err != nil {
			req.Cancel(err)
		}

		if cmd.NewTags.Len() > 0 {
			m := object.GetMetadata()
			m.SetTags(cmd.NewTags)
		}
	}

	req.Must(errors.MakeFuncContextFromFuncErr(localWorkingCopy.Lock))

	for _, pair := range pairs {
		if err := localWorkingCopy.GetStore().CreateOrUpdateDefaultProto(
			pair.object,
			sku.StoreOptions{
				MergeCheckedOut: true,
			},
		); err != nil {
			req.Cancel(err)
		}
	}

	req.Must(errors.MakeFuncContextFromFuncErr(localWorkingCopy.Unlock))
}

type externalBlobPair struct {
	objectIdString string
	pathOrSha      string

	*ids.ZettelId
	BlobFD  fd.FD
	BlobSha sha.Sha

	object *sku.Transacted
}

func (pair *externalBlobPair) SetArgs(
	objectIdString, pathOrSha string,
	envRepo env_repo.Env,
) (err error) {
	pair.BlobFD.Reset()
	pair.BlobSha.Reset()

	pair.objectIdString = objectIdString
	pair.pathOrSha = pathOrSha

	if pair.ZettelId, err = ids.MakeZettelId(pair.objectIdString); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = pair.BlobFD.SetFromPath(
		envRepo.GetCwd(),
		pathOrSha,
		envRepo.GetDefaultBlobStore(),
	); err != nil {
		if errors.IsNotExist(err) {
			if err = pair.BlobSha.Set(pair.pathOrSha); err != nil {
				err = errors.Wrap(err)
				return
			}
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (pair *externalBlobPair) GetDigest() interfaces.Digest {
	if !pair.BlobFD.IsEmpty() {
		return pair.BlobFD.GetDigest()
	} else {
		return pair.BlobSha.GetDigest()
	}
}

func (pair *externalBlobPair) PopulateBlobSha() (err error) {
	if err = pair.object.SetBlobSha(pair.GetDigest()); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
