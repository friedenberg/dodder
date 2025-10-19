package commands_madder

import (
	"io"
	"os/exec"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/delim_io"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/script_value"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/command_components_madder"
)

func init() {
	utility.AddCmd("cat", &Cat{})
}

type Cat struct {
	command_components_madder.EnvBlobStore
	command_components_madder.BlobStore

	BlobStoreIndexOrConfigPath string

	Utility   script_value.Utility
	PrefixSha bool
}

var _ interfaces.CommandComponentWriter = (*Cat)(nil)

func (cmd *Cat) SetFlagDefinitions(
	flagSet interfaces.CLIFlagDefinitions,
) {
	flagSet.Var(&cmd.Utility, "utility", "")
	flagSet.StringVar(&cmd.BlobStoreIndexOrConfigPath, "blob-store", "", "")
	flagSet.BoolVar(&cmd.PrefixSha, "prefix-sha", false, "")
}

type blobIdWithReadCloser struct {
	BlobId     interfaces.MarklId
	ReadCloser io.ReadCloser
}

func (cmd Cat) makeBlobWriter(
	envRepo env_repo.BlobStoreEnv,
	blobStore blob_stores.BlobStoreInitialized,
) interfaces.FuncIter[blobIdWithReadCloser] {
	if cmd.Utility.IsEmpty() {
		return quiter.MakeSyncSerializer(
			func(readCloser blobIdWithReadCloser) (err error) {
				if err = cmd.copy(envRepo, blobStore, readCloser); err != nil {
					err = errors.Wrap(err)
					return err
				}

				return err
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
					return err
				}

				if err = utility.Start(); err != nil {
					err = errors.Wrap(err)
					return err
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
					return err
				}

				if err = utility.Wait(); err != nil {
					err = errors.Wrap(err)
					return err
				}

				return err
			},
		)
	}
}

func (cmd Cat) Run(req command.Request) {
	envRepo := cmd.MakeEnvBlobStore(req)
	blobStore := cmd.MakeBlobStore(
		envRepo,
		cmd.BlobStoreIndexOrConfigPath,
	)

	blobWriter := cmd.makeBlobWriter(envRepo, blobStore)

	for _, blobIdString := range req.PopArgs() {
		var blobId markl.Id

		if err := markl.SetMaybeSha256(
			&blobId,
			blobIdString,
		); err != nil {
			envRepo.Cancel(err)
		}

		if err := cmd.blob(blobStore, blobId, blobWriter); err != nil {
			ui.Err().Print(err)
		}
	}
}

func (cmd Cat) copy(
	envBlobStore env_repo.BlobStoreEnv,
	blobStore blob_stores.BlobStoreInitialized,
	readCloser blobIdWithReadCloser,
) (err error) {
	defer errors.DeferredCloser(&err, readCloser.ReadCloser)

	if cmd.PrefixSha {
		if _, err = delim_io.CopyWithPrefixOnDelim(
			'\n',
			markl.FormatBytesAsHex(readCloser.BlobId),
			envBlobStore.GetUI(),
			readCloser.ReadCloser,
			true,
		); err != nil {
			err = errors.Wrap(err)
			return err
		}
	} else {
		if _, err = io.Copy(envBlobStore.GetUIFile(), readCloser.ReadCloser); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (cmd Cat) blob(
	blobStore blob_stores.BlobStoreInitialized,
	blobId interfaces.MarklId,
	blobWriter interfaces.FuncIter[blobIdWithReadCloser],
) (err error) {
	var reader interfaces.BlobReader

	if reader, err = blobStore.MakeBlobReader(blobId); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = blobWriter(blobIdWithReadCloser{BlobId: blobId, ReadCloser: reader}); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
