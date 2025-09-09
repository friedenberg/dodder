package commands

import (
	"io"
	"sync/atomic"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/markl_io"
	"code.linenisgreat.com/dodder/go/src/delta/script_value"
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

func (cmd *BlobStoreWrite) SetFlagSet(flagSet interfaces.CommandLineFlagDefinitions) {
	flagSet.BoolVar(
		&cmd.Check,
		"check",
		false,
		"only check if the object already exists",
	)

	flagSet.Var(&cmd.UtilityBefore, "utility-before", "")
	flagSet.Var(&cmd.UtilityAfter, "utility-after", "")
}

type blobWriteResult struct {
	error
	interfaces.MarklId
	Path string
}

// TODO add support for blob store ids
func (cmd BlobStoreWrite) Run(req command.Request) {
	blobStore := cmd.MakeBlobStoreLocal(
		req,
		req.Config,
		env_ui.Options{},
		local_working_copy.OptionsEmpty,
	)

	var failCount atomic.Uint32

	sawStdin := false

	for _, arg := range req.PopArgs() {
		switch {
		case sawStdin:
			ui.Err().Print("'-' passed in more than once. Ignoring")
			continue

		case arg == "-":
			sawStdin = true
		}

		result := blobWriteResult{Path: arg}

		result.MarklId, result.error = cmd.doOne(blobStore, arg)

		if result.IsNull() {
			ui.Err().Printf("digest for arg %q was null", arg)
			continue
		}

		if result.error != nil {
			blobStore.GetErr().Printf(
				"%s: (error: %q)",
				result.Path,
				result.error,
			)
			failCount.Add(1)
			continue
		}

		hasBlob := blobStore.HasBlob(result.MarklId)

		if hasBlob {
			if cmd.Check {
				blobStore.GetUI().Printf(
					"%s %s (already checked in)",
					markl.Format(result.MarklId),
					result.Path,
				)
			} else {
				blobStore.GetUI().Printf(
					"%s %s (checked in)",
					markl.Format(result.MarklId),
					result.Path,
				)
			}
		} else {
			ui.Err().Printf(
				"%s %s (untracked)",
				markl.Format(result.MarklId),
				result.Path,
			)

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

// TODO rewrite to just return blobWriteResult
func (cmd BlobStoreWrite) doOne(
	blobStore command_components.BlobStoreWithEnv,
	path string,
) (blobId interfaces.MarklId, err error) {
	var readCloser io.ReadCloser

	if readCloser, err = env_dir.NewFileReader(
		env_dir.DefaultConfig,
		path,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, readCloser)

	var writeCloser interfaces.BlobWriter

	if cmd.Check {
		{
			var repool func()
			writeCloser, repool = markl_io.MakeWriterWithRepool(
				markl.HashTypeSha256.Get(),
				nil,
			)
			defer repool()
		}
	} else {
		if writeCloser, err = blobStore.MakeBlobWriter(""); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	defer errors.DeferredCloser(&err, writeCloser)

	if _, err = io.Copy(writeCloser, readCloser); err != nil {
		err = errors.Wrap(err)
		return
	}

	blobId = writeCloser.GetMarklId()

	return
}
