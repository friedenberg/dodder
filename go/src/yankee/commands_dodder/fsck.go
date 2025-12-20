package commands_dodder

import (
	"sync/atomic"
	"time"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/kilo/command"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
	"code.linenisgreat.com/dodder/go/src/oscar/queries"
	"code.linenisgreat.com/dodder/go/src/victor/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/xray/command_components_dodder"
)

func init() {
	utility.AddCmd(
		"fsck",
		&Fsck{},
	)
}

// TODO add options to verify blobs, type formats, tags
type Fsck struct {
	command_components_dodder.LocalWorkingCopy
	command_components_dodder.InventoryLists
	command_components_dodder.Query

	InventoryListPath string
}

var _ interfaces.CommandComponentWriter = (*Fsck)(nil)

func (cmd *Fsck) SetFlagDefinitions(flagSet interfaces.CLIFlagDefinitions) {
	cmd.LocalWorkingCopy.SetFlagDefinitions(flagSet)

	flagSet.StringVar(
		&cmd.InventoryListPath,
		"inventory_list-path",
		"",
		"instead of using the store's object, verify the objects at the inventory list at the given path",
	)
}

func (cmd Fsck) Run(req command.Request) {
	repo := cmd.MakeLocalWorkingCopy(req)

	var seq interfaces.SeqError[*sku.Transacted]

	if cmd.InventoryListPath == "" {
		query := cmd.MakeQueryIncludingWorkspace(
			req,
			queries.BuilderOptions(
				queries.BuilderOptionWorkspace(repo),
				queries.BuilderOptionDefaultGenres(genres.All()...),
				queries.BuilderOptionDefaultSigil(
					ids.SigilLatest,
					ids.SigilHistory,
					ids.SigilHidden,
				),
			),
			repo,
			req.PopArgs(),
		)

		seq = repo.GetStore().All(query)

		ui.Out().Printf("verification for %q objects in progress...", query)
	} else {
		seq = cmd.MakeSeqFromPath(
			repo,
			repo.GetInventoryListCoderCloset(),
			cmd.InventoryListPath,
			nil,
		)
	}

	cmd.runVerification(repo, seq)
}

func (cmd Fsck) runVerification(
	repo *local_working_copy.Repo,
	seq interfaces.SeqError[*sku.Transacted],
) {
	var count atomic.Uint32

	type objectError struct {
		object *sku.Transacted
		err    error
	}

	var objectErrors collections_slice.Slice[objectError]

	if err := errors.RunChildContextWithPrintTicker(
		repo,
		func(ctx errors.Context) {
			for object, errIter := range seq {
				if errIter != nil {
					err := objectError{err: errors.Wrap(errIter)}

					if object != nil {
						err.object = object.CloneTransacted()
					}

					objectErrors.Append(err)

					continue
				}

				if err := markl.AssertIdIsNotNull(
					object.GetObjectDigest(),
				); err != nil {
					objectErrors.Append(
						objectError{
							err:    err,
							object: object.CloneTransacted(),
						},
					)
				}

				if err := object.Verify(); err != nil {
					objectErrors.Append(
						objectError{
							err:    err,
							object: object.CloneTransacted(),
						},
					)
				}

				count.Add(1)
			}
		},
		func(time time.Time) {
			ui.Out().Printf(
				"(in progress) %d verified, %d errors",
				count.Load(),
				len(objectErrors),
			)
		},
		3*time.Second,
	); err != nil {
		repo.Cancel(err)
		return
	}

	ui.Out().Printf("verification complete")
	ui.Out().Printf("objects verified: %d", count.Load())
	ui.Out().Printf("objects with errors: %d", len(objectErrors))

	for _, objectError := range objectErrors {
		ui.Out().Printf("%s:", sku.StringMetadataTaiMerkle(objectError.object))
		ui.CLIErrorTreeEncoder.EncodeTo(objectError.err, ui.Out())
	}
}
