package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/lima/organize_text"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
	"code.linenisgreat.com/dodder/go/src/papa/user_ops"
)

func init() {
	command.Register("clean", &Clean{})
}

type Clean struct {
	command_components.LocalWorkingCopyWithQueryGroup

	force                    bool
	includeRecognizedBlobs   bool
	includeRecognizedZettels bool
	includeParent            bool
	organize                 bool
}

var _ interfaces.CommandComponentWriter = (*Clean)(nil)

func (c *Clean) SetFlagDefinitions(f interfaces.CommandLineFlagDefinitions) {
	c.LocalWorkingCopyWithQueryGroup.SetFlagDefinitions(f)

	f.BoolVar(
		&c.force,
		"force",
		false,
		"remove objects in working directory even if they have changes",
	)

	f.BoolVar(
		&c.includeParent,
		"include-mutter",
		false,
		"remove objects in working directory if they match their Mutter",
	)

	f.BoolVar(
		&c.includeRecognizedBlobs,
		"recognized-blobs",
		false,
		"remove blobs in working directory or args that are recognized",
	)

	f.BoolVar(
		&c.includeRecognizedZettels,
		"recognized-zettelen",
		false,
		"remove Zetteln in working directory or args that are recognized",
	)

	f.BoolVar(&c.organize, "organize", false, "")
}

func (cmd Clean) Run(req command.Request) {
	localWorkingCopy, queryGroup := cmd.MakeLocalWorkingCopyAndQueryGroup(
		req,
		query.BuilderOptions(
			query.BuilderOptionHidden(nil),
			query.BuilderOptionDefaultGenres(genres.All()...),
		),
	)

	envWorkspace := localWorkingCopy.GetEnvWorkspace()
	envWorkspace.AssertNotTemporary(req)

	if cmd.organize {
		if err := cmd.runOrganize(localWorkingCopy, queryGroup); err != nil {
			localWorkingCopy.Cancel(err)
		}

		return
	}

	localWorkingCopy.Must(
		errors.MakeFuncContextFromFuncErr(localWorkingCopy.Lock),
	)

	if err := localWorkingCopy.GetStore().QuerySkuType(
		queryGroup,
		func(co sku.SkuType) (err error) {
			if !cmd.shouldClean(localWorkingCopy, co, queryGroup) {
				return err
			}

			if err = localWorkingCopy.GetStore().DeleteCheckedOut(co); err != nil {
				err = errors.Wrap(err)
				return err
			}

			return err
		},
	); err != nil {
		localWorkingCopy.Cancel(err)
	}

	localWorkingCopy.Must(
		errors.MakeFuncContextFromFuncErr(localWorkingCopy.Unlock),
	)
}

func (c Clean) runOrganize(
	u *local_working_copy.Repo,
	qg *query.Query,
) (err error) {
	opOrganize := user_ops.Organize{
		Repo: u,
		Metadata: organize_text.Metadata{
			RepoId: qg.RepoId,
			OptionCommentSet: organize_text.MakeOptionCommentSet(
				nil,
				&organize_text.OptionCommentUnknown{
					Value: "instructions: to clean an object, delete it entirely",
				},
			),
		},
		DontUseQueryGroupForOrganizeMetadata: true,
	}

	ui.Log().Print(qg)

	var organizeResults organize_text.OrganizeResults

	if organizeResults, err = opOrganize.RunWithQueryGroup(
		qg,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	var changes organize_text.Changes

	if changes, err = organize_text.ChangesFromResults(
		u.GetConfig().GetPrintOptions(),
		organizeResults,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	u.Must(errors.MakeFuncContextFromFuncErr(u.Lock))

	for _, el := range changes.Removed.AllSkuAndIndex() {
		if err = u.GetStore().DeleteCheckedOut(
			el,
		); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	u.Must(errors.MakeFuncContextFromFuncErr(u.Unlock))

	return err
}

func (cmd Clean) shouldClean(
	u *local_working_copy.Repo,
	co sku.SkuType,
	qg *query.Query,
) bool {
	if cmd.force {
		return true
	}

	state := co.GetState()

	switch state {
	case checked_out_state.CheckedOut:
		return sku.InternalAndExternalEqualsWithoutTai(co)

	case checked_out_state.Recognized:
		return !qg.ExcludeRecognized
	}

	if cmd.includeParent {
		mother := sku.GetTransactedPool().Get()
		defer sku.GetTransactedPool().Put(mother)

		err := u.GetStore().GetStreamIndex().ReadOneObjectId(
			co.GetSku().GetObjectId(),
			mother,
		)

		errors.PanicIfError(err)

		if object_metadata.EqualerSansTai.Equals(
			&co.GetSkuExternal().GetSku().Metadata,
			&mother.Metadata,
		) {
			return true
		}
	}

	return false
}
