package local_working_copy

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/lima/organize_text"
)

func (local *Repo) MakeOrganizeOptionsWithOrganizeMetadata(
	organizeFlags organize_text.Flags,
	metadata organize_text.Metadata,
) organize_text.Options {
	options := organizeFlags.GetOptions(
		local.GetConfig().GetCLIConfig().PrintOptions,
		nil,
		local.SkuFormatBoxCheckedOutNoColor(),
		local.GetStore().GetAbbrStore().GetAbbr(),
		sku.ObjectFactory{},
	)

	options.Metadata = metadata

	return options
}

func (local *Repo) MakeOrganizeOptionsWithQueryGroup(
	organizeFlags organize_text.Flags,
	qg *query.Query,
) organize_text.Options {
	return organizeFlags.GetOptions(
		local.GetConfig().GetCLIConfig().PrintOptions,
		query.GetTags(qg),
		local.SkuFormatBoxCheckedOutNoColor(),
		local.GetStore().GetAbbrStore().GetAbbr(),
		sku.ObjectFactory{},
	)
}

func (local *Repo) LockAndCommitOrganizeResults(
	results organize_text.OrganizeResults,
) (changeResults organize_text.Changes, err error) {
	if changeResults, err = organize_text.ChangesFromResults(
		local.GetConfig().GetCLIConfig().PrintOptions,
		results,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	local.Must(errors.MakeFuncContextFromFuncErr(local.Lock))

	count := changeResults.Changed.Len()

	if count > 30 {
		if !local.Confirm(
			fmt.Sprintf(
				"a large number (%d) of objects are being changed. continue to commit?",
				count,
			),
		) {
			// TODO output organize file used
			errors.ContextCancelWith499ClientClosedRequest(local)
			return
		}
	}

	var proto sku.Proto

	workspace := local.GetEnvWorkspace()
	workspaceType := workspace.GetDefaults().GetType()

	proto.Type = workspaceType

	for _, changed := range changeResults.Changed.AllSkuAndIndex() {
		if err = local.GetStore().CreateOrUpdate(
			changed.GetSkuExternal(),
			sku.CommitOptions{
				Proto: proto,
				StoreOptions: sku.StoreOptions{
					MergeCheckedOut: true,
				},
			},
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	local.Must(errors.MakeFuncContextFromFuncErr(local.Unlock))

	return
}

func (local *Repo) ApplyToOrganizeOptions(oo *organize_text.Options) {
	oo.Config = local.GetConfig()
	oo.Abbr = local.GetStore().GetAbbrStore().GetAbbr()

	if !local.GetConfig().GetCLIConfig().IsDryRun() {
		return
	}

	oo.AddPrototypeAndOption(
		"dry-run",
		&organize_text.OptionCommentDryRun{
			MutableConfigDryRun: local.GetConfig(),
		},
	)
}
