package commands

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/markl_age_id"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/papa/command_components_dodder"
)

func init() {
	utility.AddCmd(
		"export",
		&Export{
			CompressionType: compression_type.CompressionTypeEmpty,
		})

}

type Export struct {
	command_components_dodder.LocalWorkingCopyWithQueryGroup

	AgeIdentity     age.Identity
	CompressionType compression_type.CompressionType
}

var _ interfaces.CommandComponentWriter = (*Export)(nil)

func (cmd *Export) SetFlagDefinitions(f interfaces.CLIFlagDefinitions) {
	cmd.LocalWorkingCopyWithQueryGroup.SetFlagDefinitions(f)

	f.Var(&cmd.AgeIdentity, "age-identity", "")
	cmd.CompressionType.SetFlagDefinitions(f)
}

func (cmd Export) Run(req command.Request) {
	localWorkingCopy, queryGroup := cmd.MakeLocalWorkingCopyAndQueryGroup(
		req,
		query.BuilderOptions(
			query.BuilderOptionDefaultSigil(
				ids.SigilHistory,
				ids.SigilHidden,
			),
			query.BuilderOptionDefaultGenres(
				genres.InventoryList,
			),
		),
	)

	var list *sku.ListTransacted

	{
		var err error

		if list, err = localWorkingCopy.MakeInventoryList(queryGroup); err != nil {
			localWorkingCopy.Cancel(err)
		}
	}

	var ag markl_age_id.Id

	if err := ag.AddIdentity(cmd.AgeIdentity); err != nil {
		errors.ContextCancelWithErrorAndFormat(
			localWorkingCopy,
			err,
			"age-identity: %q",
			&cmd.AgeIdentity,
		)
	}

	var writeCloser io.WriteCloser = ohio.NopWriteCloser(localWorkingCopy.GetUIFile())

	defer errors.ContextMustClose(localWorkingCopy, writeCloser)

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(writeCloser)
	defer repoolBufferedWriter()
	defer errors.ContextMustFlush(localWorkingCopy, bufferedWriter)

	inventoryListCoderCloset := localWorkingCopy.GetInventoryListCoderCloset()

	if _, err := inventoryListCoderCloset.WriteTypedBlobToWriter(
		req,
		ids.GetOrPanic(localWorkingCopy.GetImmutableConfigPublic().GetInventoryListTypeId()).Type,
		quiter.MakeSeqErrorFromSeq(list.All()),
		bufferedWriter,
	); err != nil {
		localWorkingCopy.Cancel(err)
	}
}
