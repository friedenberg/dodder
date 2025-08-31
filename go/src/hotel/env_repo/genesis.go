package env_repo

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

func (env *Env) Genesis(bigBang BigBang) {
	if env.directoryLayout == nil {
		errors.ContextCancelWithErrorf(
			env,
			"directory layout not initialized",
		)
	}

	if err := bigBang.GenesisConfig.Blob.GeneratePrivateKey(); err != nil {
		env.Cancel(err)
		return
	}

	bigBang.GenesisConfig.Blob.SetInventoryListTypeId(
		bigBang.InventoryListType.String(),
	)

	env.config.Type = bigBang.GenesisConfig.Type
	env.config.Blob = bigBang.GenesisConfig.Blob

	if err := env.MakeDir(
		env.DirObjectId(),
		env.DirCache(),
		env.DirLostAndFound(),

		// TODO remove
		env.DirFirstBlobStoreInventoryLists(),
		env.DirFirstBlobStoreBlobs(),

		// TODO refactor
		env.DirBlobStores("0"),
		env.DirBlobStoreConfigs(),
	); err != nil {
		env.Cancel(err)
	}

	env.writeInventoryListLog()
	env.writeConfig(bigBang)
	env.writeBlobStoreConfig(bigBang)

	if env.config.Blob.GetRepoType() == repo_type.TypeWorkingCopy {
		if err := ohio.CopyFileLines(
			bigBang.Yin,
			filepath.Join(env.DirObjectId(), "Yin"),
		); err != nil {
			env.Cancel(err)
		}

		if err := ohio.CopyFileLines(
			bigBang.Yang,
			filepath.Join(env.DirObjectId(), "Yang"),
		); err != nil {
			env.Cancel(err)
		}

		env.writeFile(env.FileConfigMutable(), "")
		env.writeFile(env.FileCacheDormant(), "")
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
			env.Cancel(err)
		}

		defer errors.ContextMustClose(env, file)
	}

	coder := triple_hyphen_io.Coder[*triple_hyphen_io.TypedBlobEmpty]{
		Metadata: triple_hyphen_io.TypedMetadataCoder[struct{}]{},
	}

	tipe := ids.GetOrPanic(
		env.config.Blob.GetInventoryListTypeId(),
	).Type

	subject := triple_hyphen_io.TypedBlobEmpty{
		Type: tipe,
	}

	if _, err := coder.EncodeTo(&subject, file); err != nil {
		env.Cancel(err)
	}
}

func (env *Env) writeConfig(bigBang BigBang) {
	if err := triple_hyphen_io.EncodeToFile(
		genesis_configs.CoderPrivate,
		&env.config,
		env.FileConfigPermanent(),
	); err != nil {
		env.Cancel(err)
		return
	}
}

func (env *Env) writeBlobStoreConfig(bigBang BigBang) {
	if store_version.IsCurrentVersionLessOrEqualToV10() {
		// the immutable config contains the only blob stores's config
		return
	}

	if err := triple_hyphen_io.EncodeToFile(
		blob_store_configs.Coder,
		&blob_store_configs.TypedConfig{
			Type: bigBang.TypedBlobStoreConfig.Type,
			Blob: bigBang.TypedBlobStoreConfig.Blob,
		},
		env.DirBlobStoreConfigs(
			fmt.Sprintf("%d-default.%s", 0, FileNameBlobStoreConfig),
		),
	); err != nil {
		env.Cancel(err)
		return
	}
}

func (env *Env) writeFile(path string, contents any) {
	var file *os.File

	{
		var err error

		if file, err = files.CreateExclusiveWriteOnly(path); err != nil {
			if errors.IsExist(err) {
				ui.Err().Printf("%s already exists, not overwriting", path)
				err = nil
			} else {
				env.Cancel(err)
			}
		}
	}

	defer errors.ContextMustClose(env, file)

	if value, ok := contents.(string); ok {
		if _, err := io.WriteString(file, value); err != nil {
			env.Cancel(err)
		}
	} else {
		// TODO remove gob
		enc := gob.NewEncoder(file)

		if err := enc.Encode(contents); err != nil {
			env.Cancel(err)
		}
	}
}
