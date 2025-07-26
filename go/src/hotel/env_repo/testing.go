//go:build test

package env_repo

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/debug"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
)

func MakeTesting(
	t *ui.TestContext,
	contents map[string]string,
) (envRepo Env) {
	t = t.Skip(1)

	dirTemp := t.TempDir()

	envDir := env_dir.MakeWithXDGRootOverrideHomeAndInitialize(
		t.Context,
		dirTemp,
		debug.Options{},
	)

	{
		var err error

		if envRepo, err = Make(
			env_local.Make(env_ui.MakeDefault(t.Context), envDir),
			Options{
				BasePath:                dirTemp,
				PermitNoDodderDirectory: true,
			},
		); err != nil {
			errors.ContextCancelWithErrorAndFormat(
				t.Context,
				err,
				"EnvRepo: %#v",
				envRepo,
			)
		}
	}

	var bigBang BigBang

	bigBang.SetDefaults()

	envRepo.Genesis(bigBang)

	if contents == nil {
		return
	}

	for shaExpected, content := range contents {
		var writeCloser interfaces.WriteCloseBlobIdGetter

		writeCloser, err := envRepo.GetDefaultBlobStore().BlobWriter()
		if err != nil {
			errors.ContextCancelWithErrorAndFormat(
				t.Context,
				err,
				"failed to make blob writer",
			)
		}

		reader, repool := pool.GetStringReader(content)
		defer repool()
		_, err = io.Copy(writeCloser, reader)
		if err != nil {
			errors.ContextCancelWithErrorAndFormat(
				t.Context,
				err,
				"failed to write string to blob writer",
			)
		}

		err = writeCloser.Close()
		if err != nil {
			errors.ContextCancelWithErrorAndFormat(
				t.Context,
				err, "failed to write string to blob writer",
			)
		}

		shActual := writeCloser.GetBlobId()
		expected := sha.MustWithString(shaExpected)

		err = expected.AssertEqualsShaLike(shActual)
		if err != nil {
			errors.ContextCancelWithErrorAndFormat(
				t.Context,
				err, "sha mismatch: %s, %q", shaExpected, content,
			)
		}
	}

	return
}
