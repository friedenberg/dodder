//go:build test

package env_repo

import (
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/debug"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
)

func MakeTesting(
	t *ui.T,
	contents map[string]string,
) (envRepo Env) {
	t = t.Skip(1)

	ctx := errors.MakeContextDefault()

	if err := ctx.Run(
		func(ctx interfaces.Context) {
			dirTemp := t.TempDir()

			envDir := env_dir.MakeWithXDGRootOverrideHomeAndInitialize(
				ctx,
				dirTemp,
				debug.Options{},
			)

			var envRepo Env

			{
				var err error

				if envRepo, err = Make(
					env_local.Make(env_ui.MakeDefault(ctx), envDir),
					Options{
						BasePath:                dirTemp,
						PermitNoDodderDirectory: true,
					},
				); err != nil {
					errors.ContextCancelWithErrorAndFormat(ctx, err, "EnvRepo: %#v", envRepo)
				}
			}

			var bigBang BigBang

			bigBang.SetDefaults()

			envRepo.Genesis(bigBang)
			ui.Debug().Print(envRepo)

			if contents == nil {
				return
			}

			for shaExpected, content := range contents {
				var writeCloser interfaces.WriteCloseDigester

				writeCloser, err := envRepo.GetDefaultBlobStore().BlobWriter()
				if err != nil {
					errors.ContextCancelWithErrorAndFormat(
						ctx,
						err,
						"failed to make blob writer",
					)
				}

				_, err = io.Copy(writeCloser, strings.NewReader(content))
				if err != nil {
					errors.ContextCancelWithErrorAndFormat(
						ctx,
						err,
						"failed to write string to blob writer",
					)
				}

				err = writeCloser.Close()
				if err != nil {
					errors.ContextCancelWithErrorAndFormat(
						ctx,
						err, "failed to write string to blob writer",
					)
				}

				shActual := writeCloser.GetDigest()
				expected := sha.Must(shaExpected)

				err = expected.AssertEqualsShaLike(shActual)
				if err != nil {
					errors.ContextCancelWithErrorAndFormat(
						ctx,
						err, "sha mismatch: %s, %q", shaExpected, content,
					)
				}
			}
		},
	); err != nil {
		t.Fatalf("making envRepo failed: %s", err)
	}

	return
}

func MakeTesting2(
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
	ui.Debug().Print(envRepo)

	if contents == nil {
		return
	}

	for shaExpected, content := range contents {
		var writeCloser interfaces.WriteCloseDigester

		writeCloser, err := envRepo.GetDefaultBlobStore().BlobWriter()
		if err != nil {
			errors.ContextCancelWithErrorAndFormat(
				t.Context,
				err,
				"failed to make blob writer",
			)
		}

		_, err = io.Copy(writeCloser, strings.NewReader(content))
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

		shActual := writeCloser.GetDigest()
		expected := sha.Must(shaExpected)

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
