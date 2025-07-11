package env_repo

import (
	"encoding/gob"
	"io"
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_config"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io2"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
	"code.linenisgreat.com/dodder/go/src/golf/genesis_config_io"
)

func (env *Env) Genesis(bigBang BigBang) {
	if err := bigBang.GenesisConfig.GeneratePrivateKey(); err != nil {
		env.CancelWithError(err)
		return
	}

	env.config.Type = bigBang.Type
	env.config.Blob = bigBang.GenesisConfig

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
	env.writeConfig(bigBang)
	env.writeBlobStoreConfig(bigBang)

	if env.config.Blob.GetRepoType() == repo_type.TypeWorkingCopy {
		if err := ohio.CopyFileLines(
			bigBang.Yin,
			filepath.Join(env.DirObjectId(), "Yin"),
		); err != nil {
			env.CancelWithError(err)
		}

		if err := ohio.CopyFileLines(
			bigBang.Yang,
			filepath.Join(env.DirObjectId(), "Yang"),
		); err != nil {
			env.CancelWithError(err)
		}

		env.writeFile(env.FileConfigMutable(), "")
		env.writeFile(env.FileCacheDormant(), "")
	}

	env.setupStores()
}

// TODO determine if this is necessary, it appears to be writing an empty
// inventory list
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

func (env *Env) writeConfig(bigBang BigBang) {
	triple_hyphen_io2.EncodeToFile[genesis_config.Private](
		env,
		genesis_config_io.CoderPrivate,
		&env.config,
		env.FileConfigPermanent(),
	)
}

func (env *Env) writeBlobStoreConfig(bigBang BigBang) {
	if store_version.LessOrEqual(
		bigBang.GenesisConfig.GetStoreVersion(),
		store_version.V10,
	) {
		// the immutable config contains the only blob stores's config
		return
	}

	// TODO write blob store config
}

// TODO remove gob
func (env *Env) writeFile(path string, contents any) {
	var file *os.File

	{
		var err error

		if file, err = files.CreateExclusiveWriteOnly(path); err != nil {
			if errors.IsExist(err) {
				ui.Err().Printf("%s already exists, not overwriting", path)
				err = nil
			} else {
				env.CancelWithError(err)
			}
		}
	}

	defer env.MustClose(file)

	if value, ok := contents.(string); ok {
		if _, err := io.WriteString(file, value); err != nil {
			env.CancelWithError(err)
		}
	} else {
		enc := gob.NewEncoder(file)

		if err := enc.Encode(contents); err != nil {
			env.CancelWithError(err)
		}
	}
}
