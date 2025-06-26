package commands

import (
	"flag"
	"io"
	"os/exec"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/delim_io"
	"code.linenisgreat.com/dodder/go/src/delta/script_value"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("cat-blob", &CatBlob{})
}

type CatBlob struct {
	command_components.BlobStoreLocal

	Utility   script_value.Utility
	PrefixSha bool
}

func (cmd *CatBlob) SetFlagSet(f *flag.FlagSet) {
	f.Var(&cmd.Utility, "utility", "")
	f.BoolVar(&cmd.PrefixSha, "prefix-sha", false, "")
}

type shaWithReadCloser struct {
	Sha        *sha.Sha
	ReadCloser io.ReadCloser
}

func (c CatBlob) makeBlobWriter(
	blobStore command_components.BlobStoreWithEnv,
) interfaces.FuncIter[shaWithReadCloser] {
	if c.Utility.IsEmpty() {
		return quiter.MakeSyncSerializer(
			func(rc shaWithReadCloser) (err error) {
				if err = c.copy(blobStore, rc); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			},
		)
	} else {
		return quiter.MakeSyncSerializer(
			func(rc shaWithReadCloser) (err error) {
				defer errors.DeferredCloser(&err, rc.ReadCloser)

				cmd := exec.Command(c.Utility.Head(), c.Utility.Tail()...)
				cmd.Stdin = rc.ReadCloser

				var out io.ReadCloser

				if out, err = cmd.StdoutPipe(); err != nil {
					err = errors.Wrap(err)
					return
				}

				if err = cmd.Start(); err != nil {
					err = errors.Wrap(err)
					return
				}

				if err = c.copy(
					blobStore,
					shaWithReadCloser{
						Sha:        rc.Sha,
						ReadCloser: out,
					},
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				if err = cmd.Wait(); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			},
		)
	}
}

func (cmd CatBlob) Run(
	dep command.Request,
) {
	blobStore := cmd.MakeBlobStoreLocal(
		dep,
		dep.Config,
		env_ui.Options{},
		local_working_copy.OptionsEmpty,
	)

	blobWriter := cmd.makeBlobWriter(blobStore)

	for _, v := range dep.PopArgs() {
		var sh sha.Sha

		if err := sh.Set(v); err != nil {
			blobStore.CancelWithError(err)
		}

		if err := cmd.blob(blobStore, &sh, blobWriter); err != nil {
			ui.Err().Print(err)
		}
	}
}

func (c CatBlob) copy(
	blobStore command_components.BlobStoreWithEnv,
	rc shaWithReadCloser,
) (err error) {
	defer errors.DeferredCloser(&err, rc.ReadCloser)

	if c.PrefixSha {
		if _, err = delim_io.CopyWithPrefixOnDelim(
			'\n',
			rc.Sha.GetShaLike().GetShaString(),
			blobStore.GetUI(),
			rc.ReadCloser,
			true,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		if _, err = io.Copy(blobStore.GetUIFile(), rc.ReadCloser); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (c CatBlob) blob(
	blobStore command_components.BlobStoreWithEnv,
	sh *sha.Sha,
	blobWriter interfaces.FuncIter[shaWithReadCloser],
) (err error) {
	var r sha.ReadCloser

	if r, err = blobStore.BlobReader(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = blobWriter(shaWithReadCloser{Sha: sh, ReadCloser: r}); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
