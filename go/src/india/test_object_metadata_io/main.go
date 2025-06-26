package test_object_metadata_io

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
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

func Make(
	t *ui.T,
	contents map[string]string,
) (f env_repo.Env) {
	t = t.Skip(1)

	p := t.TempDir()

	dir := env_dir.MakeWithHome(
		errors.MakeContextDefault(),
		p,
		debug.Options{},
		false,
	)

	var err error

	if f, err = env_repo.Make(
		env_local.Make(env_ui.MakeDefault(), dir),
		env_repo.Options{
			BasePath:             p,
			PermitNoDodderDirectory: true,
		},
	); err != nil {
		t.Fatalf("failed to make dir_layout: %s", err)
	}

	if contents == nil {
		return
	}

	for k, e := range contents {
		var w sha.WriteCloser

		w, err := f.BlobWriter()
		if err != nil {
			t.Fatalf("failed to make blob writer: %s", err)
		}

		_, err = io.Copy(w, strings.NewReader(e))
		if err != nil {
			t.Fatalf("failed to write string to blob writer: %s", err)
		}

		err = w.Close()
		if err != nil {
			t.Fatalf("failed to write string to blob writer: %s", err)
		}

		sh := w.GetShaLike()
		expected := sha.Must(k)

		err = expected.AssertEqualsShaLike(sh)
		if err != nil {
			t.Fatalf("sha mismatch: %s. %s, %q", err, k, e)
		}
	}

	return
}
