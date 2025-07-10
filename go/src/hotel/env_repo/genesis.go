package env_repo

import (
	"bufio"
	"encoding/gob"
	"io"
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_config"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io2"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
	"code.linenisgreat.com/dodder/go/src/golf/genesis_config_io"
)

func (env *Env) Genesis(bb BigBang) {
	if err := bb.Config.GeneratePrivateKey(); err != nil {
		env.CancelWithError(err)
		return
	}

	env.config.Type = bb.Type
	env.config.Blob = bb.Config

	if err := env.MakeDir(
		env.DirObjectId(),
		env.DirCache(),
		env.DirLostAndFound(),
		env.DirFirstBlobStoreInventoryLists(),
		env.DirFirstBlobStoreBlobs(),
	); err != nil {
		env.CancelWithError(err)
	}

	env.writeInventoryListLog()

	{
		var f *os.File

		{
			var err error

			if f, err = files.CreateExclusiveWriteOnly(
				env.FileConfigPermanent(),
			); err != nil {
				env.CancelWithError(err)
			}

			defer env.MustClose(f)
		}

		encoder := genesis_config_io.CoderPrivate{}

		if _, err := encoder.EncodeTo(&env.config, f); err != nil {
			env.CancelWithError(err)
		}
	}

	env.writeBlobStoreConfig(bb.Config)

	if env.config.Blob.GetRepoType() == repo_type.TypeWorkingCopy {
		if err := env.readAndTransferLines(
			bb.Yin,
			filepath.Join(env.DirObjectId(), "Yin"),
		); err != nil {
			env.CancelWithError(err)
		}

		if err := env.readAndTransferLines(
			bb.Yang,
			filepath.Join(env.DirObjectId(), "Yang"),
		); err != nil {
			env.CancelWithError(err)
		}

		writeFile(env.FileConfigMutable(), "")
		writeFile(env.FileCacheDormant(), "")
	}

	env.setupStores()
}

func (env Env) writeInventoryListLog() {
	var file *os.File

	{
		var err error

		if file, err = files.CreateExclusiveWriteOnly(
			env.FileInventoryListLog(),
		); err != nil {
			env.CancelWithError(err)
		}

		defer env.MustClose(file)
	}

	coder := triple_hyphen_io2.Coder[*triple_hyphen_io2.TypedBlobEmpty]{
		Metadata: triple_hyphen_io2.TypedMetadataCoder[struct{}]{},
	}

	tipe := builtin_types.GetOrPanic(
		builtin_types.InventoryListTypeVCurrent,
	).Type

	subject := triple_hyphen_io2.TypedBlobEmpty{
		Type: tipe,
	}

	if _, err := coder.EncodeTo(&subject, file); err != nil {
		env.CancelWithError(err)
	}
}

func (env Env) readAndTransferLines(in, out string) (err error) {
	if in == "" {
		return
	}

	var fi, fo *os.File

	if fi, err = files.Open(in); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.Deferred(&err, fi.Close)

	if fo, err = files.CreateExclusiveWriteOnly(out); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.Deferred(&err, fo.Close)

	r := bufio.NewReader(fi)
	w := bufio.NewWriter(fo)

	defer errors.Deferred(&err, w.Flush)

	for {
		var l string
		l, err = r.ReadString('\n')

		if errors.Is(err, io.EOF) {
			err = nil
			break
		}

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		// TODO-P2 sterilize line
		w.WriteString(l)
	}

	return
}

func (env *Env) writeBlobStoreConfig(config genesis_config.Private) {
	// TODO
}

func writeFile(p string, contents any) {
	var f *os.File
	var err error

	if f, err = files.CreateExclusiveWriteOnly(p); err != nil {
		if errors.IsExist(err) {
			ui.Err().Printf("%s already exists, not overwriting", p)
			err = nil
		} else {
		}

		return
	}

	defer errors.PanicIfError(err)
	defer errors.DeferredCloser(&err, f)

	if s, ok := contents.(string); ok {
		_, err = io.WriteString(f, s)
	} else {
		enc := gob.NewEncoder(f)
		err = enc.Encode(contents)
	}
}
