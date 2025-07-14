package env_repo

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

func (env Env) ReadAllLevel2Files(
	p string,
	w interfaces.FuncIter[string],
) (err error) {
	if err = files.ReadDirNamesLevel2(
		files.MakeDirNameWriterIgnoringHidden(w),
		p,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO transform into interfaces.SeqError
func (env Env) ReadAllShas(
	p string,
	w interfaces.FuncIter[*sha.Sha],
) (err error) {
	wf := func(p string) (err error) {
		var sh *sha.Sha

		if sh, err = sha.MakeShaFromPath(p); err != nil {
			ui.Err().Printf("invalid format: %q", p)
			err = nil
			return
		}

		if err = w(sh); err != nil {
			err = errors.Wrapf(err, "Sha: %s", sh)
			return
		}

		return
	}

	if err = env.ReadAllLevel2Files(p, wf); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO transform into interfaces.SeqError
func (env Env) ReadAllShasForBlobs(
	w interfaces.FuncIter[*sha.Sha],
) (err error) {
	p := env.DirFirstBlobStoreBlobs()

	if err = env.ReadAllShas(p, w); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
