package commands

import (
	"sync"
	"sync/atomic"
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	pkg_query "code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register(
		"fsck",
		&Fsck{},
	)
}

// TODO add options to verify blobs, type formats, tags
type Fsck struct {
	command_components.LocalWorkingCopy
	command_components.Query
}
var _ interfaces.CommandComponentWriter = (*Fsck)(nil)

func (cmd *Fsck) SetFlagDefinitions(flagSet interfaces.CommandLineFlagDefinitions) {
	cmd.LocalWorkingCopy.SetFlagDefinitions(flagSet)
}

func (cmd Fsck) Run(req command.Request) {
	repo := cmd.MakeLocalWorkingCopy(req)

	query := cmd.MakeQueryIncludingWorkspace(
		req,
		pkg_query.BuilderOptions(
			pkg_query.BuilderOptionWorkspace(repo),
			pkg_query.BuilderOptionDefaultGenres(genres.All()...),
			pkg_query.BuilderOptionDefaultSigil(
				ids.SigilLatest,
				ids.SigilHistory,
				ids.SigilHidden,
			),
		),
		repo,
		req.PopArgs(),
	)

	ui.Out().Printf("verification for %q objects in progress...", query)

	var count atomic.Uint32

	type objectError struct {
		object *sku.Transacted
		err    error
	}

	var objectErrorsLock sync.Mutex
	var objectErrors []objectError

	if err := errors.RunChildContextWithPrintTicker(
		repo,
		func(ctx interfaces.Context) {
			if err := repo.GetStore().QueryTransacted(
				query,
				func(object *sku.Transacted) (err error) {
					if err = markl.AssertIdIsNotNull(
						object.GetObjectDigest(),
						"object-dig",
					); err != nil {
						objectErrorsLock.Lock()

						objectErrors = append(
							objectErrors,
							objectError{
								err:    err,
								object: object.CloneTransacted(),
							},
						)

						objectErrorsLock.Unlock()

						err = nil
						return
					}

					if err = object.Verify(); err != nil {
						objectErrorsLock.Lock()

						objectErrors = append(
							objectErrors,
							objectError{
								err:    err,
								object: object.CloneTransacted(),
							},
						)

						objectErrorsLock.Unlock()

						err = nil
						return
					}

					count.Add(1)

					return
				},
			); err != nil {
				ui.Err().Print(err)
				err = nil
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
		ui.Out().Printf(
			"%s: %s",
			sku.StringMetadataTaiMerkle(objectError.object),
			objectError.err,
		)
	}
}
