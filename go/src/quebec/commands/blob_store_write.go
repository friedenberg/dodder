package commands

import (
	"flag"
	"io"
	"sync/atomic"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/script_value"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("write-blob", &BlobStoreWrite{})
	command.Register("blob_store-write", &BlobStoreWrite{})
}

type BlobStoreWrite struct {
	command_components.BlobStoreLocal

	Check         bool
	UtilityBefore script_value.Utility
	UtilityAfter  script_value.Utility
}

func (cmd *BlobStoreWrite) SetFlagSet(flagSet *flag.FlagSet) {
	flagSet.BoolVar(
		&cmd.Check,
		"check",
		false,
		"only check if the object already exists",
	)

	flagSet.Var(&cmd.UtilityBefore, "utility-before", "")
	flagSet.Var(&cmd.UtilityAfter, "utility-after", "")
}

type answer struct {
	error
	interfaces.Digest
	Path string
}

func (cmd BlobStoreWrite) Run(
	dep command.Request,
) {
	blobStore := cmd.MakeBlobStoreLocal(
		dep,
		dep.Blob,
		env_ui.Options{},
		local_working_copy.OptionsEmpty,
	)

	var failCount atomic.Uint32

	sawStdin := false

	for _, p := range dep.PopArgs() {
		switch {
		case sawStdin:
			ui.Err().Print("'-' passed in more than once. Ignoring")
			continue

		case p == "-":
			sawStdin = true
		}

		a := answer{Path: p}

		a.Digest, a.error = cmd.doOne(blobStore, p)

		if a.error != nil {
			blobStore.GetErr().Printf("%s: (error: %q)", a.Path, a.error)
			failCount.Add(1)
			continue
		}

		hasBlob := blobStore.HasBlob(a.Digest)

		if hasBlob {
			if cmd.Check {
				blobStore.GetUI().Printf(
					"%s %s (already checked in)",
					a.GetDigest(),
					a.Path,
				)
			} else {
				blobStore.GetUI().Printf("%s %s (checked in)", a.GetDigest(), a.Path)
			}
		} else {
			ui.Err().Printf("%s %s (untracked)", a.GetDigest(), a.Path)

			if cmd.Check {
				failCount.Add(1)
			}
		}
	}

	fc := failCount.Load()

	if fc > 0 {
		errors.ContextCancelWithBadRequestf(
			blobStore,
			"untracked objects: %d",
			fc,
		)
		return
	}
}

func (cmd BlobStoreWrite) doOne(
	blobStore command_components.BlobStoreWithEnv,
	path string,
) (sh interfaces.Digest, err error) {
	var readCloser io.ReadCloser

	if readCloser, err = env_dir.NewFileReader(
		env_dir.DefaultConfig,
		path,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, readCloser)

	var writeCloser interfaces.WriteCloseDigester

	if cmd.Check {
		writeCloser = sha.MakeWriter(sha.Env{}, nil)
	} else {
		if writeCloser, err = blobStore.BlobWriter(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	defer errors.DeferredCloser(&err, writeCloser)

	if _, err = io.Copy(writeCloser, readCloser); err != nil {
		err = errors.Wrap(err)
		return
	}

	sh = writeCloser.GetDigest()

	return
}
