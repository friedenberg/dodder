package commands

import (
	"io"
	"os/exec"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/delim_io"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
	"code.linenisgreat.com/dodder/go/src/delta/script_value"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("blob_store-cat", &BlobStoreCat{})
}

type BlobStoreCat struct {
	command_components.EnvRepo
	command_components.BlobStore

	BlobStoreIndexOrConfigPath string

	Utility   script_value.Utility
	PrefixSha bool
}

func (cmd *BlobStoreCat) SetFlagSet(flagSet *flags.FlagSet) {
	flagSet.Var(&cmd.Utility, "utility", "")
	flagSet.StringVar(&cmd.BlobStoreIndexOrConfigPath, "blob-store", "", "")
	flagSet.BoolVar(&cmd.PrefixSha, "prefix-sha", false, "")
}

type blobIdWithReadCloser struct {
	BlobId     interfaces.BlobId
	ReadCloser io.ReadCloser
}

func (cmd BlobStoreCat) makeBlobWriter(
	envRepo env_repo.Env,
	blobStore blob_stores.BlobStoreInitialized,
) interfaces.FuncIter[blobIdWithReadCloser] {
	if cmd.Utility.IsEmpty() {
		return quiter.MakeSyncSerializer(
			func(readCloser blobIdWithReadCloser) (err error) {
				if err = cmd.copy(envRepo, blobStore, readCloser); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			},
		)
	} else {
		return quiter.MakeSyncSerializer(
			func(readCloser blobIdWithReadCloser) (err error) {
				defer errors.DeferredCloser(&err, readCloser.ReadCloser)

				utility := exec.Command(cmd.Utility.Head(), cmd.Utility.Tail()...)
				utility.Stdin = readCloser.ReadCloser

				var out io.ReadCloser

				if out, err = utility.StdoutPipe(); err != nil {
					err = errors.Wrap(err)
					return
				}

				if err = utility.Start(); err != nil {
					err = errors.Wrap(err)
					return
				}

				if err = cmd.copy(
					envRepo,
					blobStore,
					blobIdWithReadCloser{
						BlobId:     readCloser.BlobId,
						ReadCloser: out,
					},
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				if err = utility.Wait(); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			},
		)
	}
}

func (cmd BlobStoreCat) Run(req command.Request) {
	envRepo := cmd.MakeEnvRepo(req, false)
	blobStore := cmd.MakeBlobStore(
		envRepo,
		cmd.BlobStoreIndexOrConfigPath,
	)

	blobWriter := cmd.makeBlobWriter(envRepo, blobStore)

	for _, v := range req.PopArgs() {
		var blobId merkle.Id

		if err := blobId.SetMaybeSha256(v); err != nil {
			envRepo.Cancel(err)
		}

		if err := cmd.blob(blobStore, blobId, blobWriter); err != nil {
			ui.Err().Print(err)
		}
	}
}

func (cmd BlobStoreCat) copy(
	envRepo env_repo.Env,
	blobStore blob_stores.BlobStoreInitialized,
	readCloser blobIdWithReadCloser,
) (err error) {
	defer errors.DeferredCloser(&err, readCloser.ReadCloser)

	if cmd.PrefixSha {
		if _, err = delim_io.CopyWithPrefixOnDelim(
			'\n',
			merkle.Format(readCloser.BlobId),
			envRepo.GetUI(),
			readCloser.ReadCloser,
			true,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		if _, err = io.Copy(envRepo.GetUIFile(), readCloser.ReadCloser); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (cmd BlobStoreCat) blob(
	blobStore blob_stores.BlobStoreInitialized,
	blobId interfaces.BlobId,
	blobWriter interfaces.FuncIter[blobIdWithReadCloser],
) (err error) {
	var r interfaces.ReadCloseBlobIdGetter

	if r, err = blobStore.BlobReader(blobId); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = blobWriter(blobIdWithReadCloser{BlobId: blobId, ReadCloser: r}); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
