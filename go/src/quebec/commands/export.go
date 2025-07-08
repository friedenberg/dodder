package commands

import (
	"flag"
	"io"

	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
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

func (cmd Export) Run(dep command.Request) {
	localWorkingCopy, queryGroup := cmd.MakeLocalWorkingCopyAndQueryGroup(
		dep,
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
			localWorkingCopy.CancelWithError(err)
		}
	}

	var ag age.Age

	if err := ag.AddIdentity(cmd.AgeIdentity); err != nil {
		localWorkingCopy.CancelWithErrorAndFormat(
			err,
			"age-identity: %q",
			&cmd.AgeIdentity,
		)
	}

	var wc io.WriteCloser

	o := env_dir.WriteOptions{
		Config: env_dir.MakeConfig(
			&cmd.CompressionType,
			&ag,
			false,
		),
		Writer: localWorkingCopy.GetUIFile(),
	}

	{
		var err error

		if wc, err = env_dir.NewWriter(o); err != nil {
			localWorkingCopy.CancelWithError(err)
		}
	}

	defer localWorkingCopy.MustClose(wc)

	bufferedWriter := ohio.BufferedWriter(wc)
	defer pool.GetBufioWriter().Put(bufferedWriter)
	defer localWorkingCopy.MustFlush(bufferedWriter)

	listFormat := localWorkingCopy.GetStore().GetInventoryListStore().FormatForVersion(
		localWorkingCopy.GetConfig().GetStoreVersion(),
	)

	if _, err := listFormat.WriteInventoryListBlob(
		list,
		bufferedWriter,
	); err != nil {
		localWorkingCopy.CancelWithError(err)
	}
}
