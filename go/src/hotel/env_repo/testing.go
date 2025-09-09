//go:build test

package env_repo

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/debug"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
)

func MakeTesting(
	t *ui.TestContext,
	contents map[string]string,
) (envRepo Env) {
	var bigBang BigBang
	bigBang.SetDefaults()

	return MakeTestingWithBigBang(t, contents, bigBang)
}

func MakeTestingWithBigBang(
	t *ui.TestContext,
	contents map[string]string,
	bigBang BigBang,
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
			t.Errorf("failed to make repo: %s", err)
			return
		}
	}

	envRepo.Genesis(bigBang)

	if contents == nil {
		return
	}

	for expectedDigestString, content := range contents {
		var writeCloser interfaces.BlobWriter

		writeCloser, err := envRepo.GetDefaultBlobStore().MakeBlobWriter("")
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

		actual := writeCloser.GetMarklId()
		expected, _ := markl.HashTypeSha256.GetBlobIdForHexString(
			expectedDigestString,
		)

		t.AssertNoError(markl.MakeErrNotEqual(expected, actual))
	}

	return
}
