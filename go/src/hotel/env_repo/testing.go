//go:build test

package env_repo

import (
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
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
		func(ctx errors.Context) {
			dirTemp := t.TempDir()

			envDir := env_dir.MakeWithHome(
				ctx,
				dirTemp,
				debug.Options{
					NoTempDirCleanup: true,
				},
				false,
			)

			var err error

			if envRepo, err = Make(
				env_local.Make(env_ui.MakeDefault(ctx), envDir),
				Options{
					BasePath:                dirTemp,
					PermitNoDodderDirectory: true,
				},
			); err != nil {
				t.Fatalf("failed to make envRepo: %s", err)
			}

			var bigBang BigBang

			bigBang.SetDefaults()
			envRepo.Genesis(bigBang)

			if contents == nil {
				return
			}

			for shaExpected, content := range contents {
				var writeCloser sha.WriteCloser

				writeCloser, err := envRepo.GetDefaultBlobStore().BlobWriter()
				if err != nil {
					t.Fatalf("failed to make blob writer: %s", err)
				}

				_, err = io.Copy(writeCloser, strings.NewReader(content))
				if err != nil {
					t.Fatalf("failed to write string to blob writer: %s", err)
				}

				err = writeCloser.Close()
				if err != nil {
					t.Fatalf("failed to write string to blob writer: %s", err)
				}

				shActual := writeCloser.GetShaLike()
				expected := sha.Must(shaExpected)

				err = expected.AssertEqualsShaLike(shActual)
				if err != nil {
					t.Fatalf("sha mismatch: %s. %s, %q", err, shaExpected, content)
				}
			}
		},
	); err != nil {
		t.Errorf("making envRepo failed: %s", err)
	}

	return
}
