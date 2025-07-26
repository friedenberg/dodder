package commands

import (
	"flag"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/inventory_list_coders"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register(
		"export",
		&Export{
			CompressionType: compression_type.CompressionTypeEmpty,
		},
	)
}

type Export struct {
	command_components.LocalWorkingCopyWithQueryGroup

	AgeIdentity     age.Identity
	CompressionType compression_type.CompressionType
}

func (cmd *Export) SetFlagSet(f *flag.FlagSet) {
	cmd.LocalWorkingCopyWithQueryGroup.SetFlagSet(f)

	f.Var(&cmd.AgeIdentity, "age-identity", "")
	cmd.CompressionType.SetFlagSet(f)
}

func (cmd Export) Run(req command.Request) {
	localWorkingCopy, queryGroup := cmd.MakeLocalWorkingCopyAndQueryGroup(
		req,
		query.BuilderOptionsOld(
			cmd,
			query.BuilderOptionDefaultSigil(
				ids.SigilHistory,
				ids.SigilHidden,
			),
			query.BuilderOptionDefaultGenres(
				genres.InventoryList,
			),
		),
	)

	var list *sku.List

	{
		var err error

		if list, err = localWorkingCopy.MakeInventoryList(queryGroup); err != nil {
			localWorkingCopy.Cancel(err)
		}
	}

	var ag age.Age

	if err := ag.AddIdentity(cmd.AgeIdentity); err != nil {
		errors.ContextCancelWithErrorAndFormat(
			localWorkingCopy,
			err,
			"age-identity: %q",
			&cmd.AgeIdentity,
		)
	}

	var writeCloser io.WriteCloser

	{
		var err error

		if writeCloser, err = env_dir.NewWriter(
			env_dir.MakeConfig(
				// TODO read from config
				sha.Env{},
				env_dir.MakeHashBucketPathJoinFunc([]int{2}),
				&cmd.CompressionType,
				&ag,
			),
			localWorkingCopy.GetUIFile(),
		); err != nil {
			localWorkingCopy.Cancel(err)
		}
	}

	defer errors.ContextMustClose(localWorkingCopy, writeCloser)

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(writeCloser)
	defer repoolBufferedWriter()
	defer errors.ContextMustFlush(localWorkingCopy, bufferedWriter)

	listFormat := localWorkingCopy.GetStore().GetInventoryListStore().FormatForVersion(
		localWorkingCopy.GetConfig().GetStoreVersion(),
	)

	if _, err := inventory_list_coders.WriteInventoryList(
		req,
		listFormat,
		quiter.MakeSeqErrorFromSeq(list.All()),
		bufferedWriter,
	); err != nil {
		localWorkingCopy.Cancel(err)
	}
}
