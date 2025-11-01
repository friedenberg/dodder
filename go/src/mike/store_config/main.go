package store_config

import (
	"encoding/gob"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/delta/file_extensions"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/env_workspace"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
)

// TODO remove gob entirely and cleanly separate store functionality (mutation,
// changes, loading and unloading) from config functionality (reading
// properties)
func init() {
	gob.Register(
		collections_value.MakeMutableValueSet[values.String](
			nil,
		),
	)

	gob.Register(
		collections_value.MakeValueSet[values.String](
			nil,
		),
	)

	gob.Register(quiter.StringerKeyer[values.String]{})
	gob.Register(quiter.StringerKeyerPtr[ids.Type, *ids.Type]{})
}

type (
	ApproximatedType = typed_blob_store.ApproximatedType

	Store interface {
		GetConfig() Config
		GetConfigPtr() *Config
		HasChanges() (ok bool)
		GetChanges() (out []string)
	}

	StoreMutable interface {
		Store

		AddTransacted(
			child *sku.Transacted,
			parent *sku.Transacted,
		) (err error)

		Initialize(
			dirLayout env_repo.Env,
			kcli repo_config_cli.Config,
		) (err error)

		Reset() error

		Flush(
			dirLayout env_repo.Env,
			blobStore typed_blob_store.Stores,
			printerHeader interfaces.FuncIter[string],
		) (err error)
	}
)

func Make() StoreMutable {
	return &store{}
}

type store struct {
	envRepo      env_repo.Env
	envWorkspace env_workspace.Env
	config       Config
}

func (store *store) GetConfig() Config {
	return store.config
}

func (store *store) GetConfigPtr() *Config {
	return &store.config
}

func (store *store) Reset() error {
	if store.config.compiled == nil {
		store.config.compiled = &compiled{}
	}

	store.config.configRepo = repo_configs.V1{}
	store.config.ExtensionsToTypes = make(map[string]string)
	store.config.TypesToExtensions = make(map[string]string)

	store.config.Tags = collections_value.MakeMutableValueSet[*tag](nil)
	store.config.InlineTypes = collections_value.MakeMutableValueSet[values.String](
		nil,
	)
	store.config.ImplicitTags = make(implicitTagMap)
	store.config.Repos = sku.MakeTransactedMutableSet()
	store.config.Types = sku.MakeTransactedMutableSet()

	sku.TransactedResetter.Reset(&store.config.Sku)

	return nil
}

func (store *store) Initialize(
	envRepo env_repo.Env,
	cli repo_config_cli.Config,
) (err error) {
	store.envRepo = envRepo
	store.config.CLI = cli
	store.Reset()
	store.config.configGenesis = envRepo.GetConfigPrivate().Blob

	errorWaitGroup := errors.MakeWaitGroupParallel()
	errorWaitGroup.Do(func() (err error) {
		if err = store.loadMutableConfig(envRepo); err != nil {
			if errors.IsNotExist(err) {
				err = nil
			} else {
				err = errors.Wrap(err)
			}
		}

		return err
	})

	if err = errorWaitGroup.GetError(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	store.config.FileExtensions = file_extensions.MakeDefaultConfig(
		store.config,
	)

	store.config.PrintOptions = options_print.MakeDefaultConfig(
		store.config.configRepo,
		store.config.CLI,
	)

	return err
}

func (store *store) AddTransacted(
	daughter *sku.Transacted,
	mother *sku.Transacted,
) (err error) {
	didChange := false

	genre := daughter.ObjectId.GetGenre()

	// if strings.Contains(daughter.GetObjectId().String(), "dodder") {
	// 	err = errors.Errorf(
	// 		"dodder tag: %q",
	// 		sku.StringMetadataSansTaiMerkle(daughter),
	// 	)

	// 	return
	// }

	switch genre {
	case genres.Type:
		if didChange, err = store.config.addType(daughter); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if didChange {
			store.config.SetNeedsRecompile(
				fmt.Sprintf("modified type: %s", daughter),
			)
		}

		return err

	case genres.Tag:
		if didChange, err = store.config.addTag(daughter, mother); err != nil {
			err = errors.Wrap(err)
			return err
		}

		var tag ids.Tag

		if err = tag.TodoSetFromObjectId(daughter.GetObjectId()); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if daughter.Metadata.GetTags().Len() > 0 {
			store.config.SetNeedsRecompile(
				fmt.Sprintf(
					"tag with tags added: %q -> %q",
					tag,
					quiter.SortedValues(daughter.Metadata.GetTags()),
				),
			)
		}

	case genres.Repo:
		if didChange, err = store.config.addRepo(daughter); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case genres.Config:
		if didChange, err = store.setTransacted(daughter); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	if genre != genres.Tag {
		return err
	}

	if !didChange {
		return err
	}

	if mother == nil {
		return err
	}

	if quiter.SetEquals(daughter.Metadata.Tags, mother.Metadata.Tags) {
		return err
	}

	store.config.SetNeedsRecompile(fmt.Sprintf("modified: %s", daughter))

	return err
}

func (store *store) setTransacted(
	object *sku.Transacted,
) (didChange bool, err error) {
	if !sku.TransactedLessor.LessPtr(&store.config.Sku, object) {
		return didChange, err
	}

	store.config.lock.Lock()
	defer store.config.lock.Unlock()

	didChange = true

	sku.Resetter.ResetWith(&store.config.Sku, object)

	store.config.setNeedsRecompile(
		fmt.Sprintf("updated konfig: %s", &store.config.Sku),
	)

	if err = store.loadMutableConfigBlob(
		store.config.Sku.GetType(),
		store.config.Sku.GetBlobDigest(),
	); err != nil {
		err = errors.Wrap(err)
		return didChange, err
	}

	return didChange, err
}
