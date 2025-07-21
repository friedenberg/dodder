package store_config

import (
	"encoding/gob"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
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
			kcli repo_config_cli.Blob,
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
	envRepo env_repo.Env
	config  Config
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
	cli repo_config_cli.Blob,
) (err error) {
	store.envRepo = envRepo
	store.config.CLI = cli
	store.Reset()
	store.config.configGenesis = envRepo.GetConfigPrivate().Blob

	wg := errors.MakeWaitGroupParallel()
	wg.Do(func() (err error) {
		if err = store.loadMutableConfig(envRepo); err != nil {
			if errors.IsNotExist(err) {
				err = nil
			} else {
				err = errors.Wrap(err)
			}
		}

		return
	})

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	store.config.ApplyPrintOptionsConfig(
		store.config.GetPrintOptions(),
	)

	return
}

func (store *store) AddTransacted(
	child *sku.Transacted,
	parent *sku.Transacted,
) (err error) {
	didChange := false

	g := child.ObjectId.GetGenre()

	switch g {
	case genres.Type:
		if didChange, err = store.config.addType(child); err != nil {
			err = errors.Wrap(err)
			return
		}

		if didChange {
			store.config.SetNeedsRecompile(
				fmt.Sprintf("modified type: %s", child),
			)
		}

		return

	case genres.Tag:
		if didChange, err = store.config.addTag(child, parent); err != nil {
			err = errors.Wrap(err)
			return
		}

		var tag ids.Tag

		if err = tag.TodoSetFromObjectId(child.GetObjectId()); err != nil {
			err = errors.Wrap(err)
			return
		}

		if child.Metadata.GetTags().Len() > 0 {
			store.config.SetNeedsRecompile(
				fmt.Sprintf(
					"tag with tags added: %q -> %q",
					tag,
					quiter.SortedValues(child.Metadata.GetTags()),
				),
			)
		}

	case genres.Repo:
		if didChange, err = store.config.addRepo(child); err != nil {
			err = errors.Wrap(err)
			return
		}

	case genres.Config:
		if didChange, err = store.setTransacted(child); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if g != genres.Tag {
		return
	}

	if !didChange {
		return
	}

	if parent == nil {
		return
	}

	if quiter.SetEquals(child.Metadata.Tags, parent.Metadata.Tags) {
		return
	}

	store.config.SetNeedsRecompile(fmt.Sprintf("modified: %s", child))

	return
}

func (store *store) setTransacted(
	kt1 *sku.Transacted,
) (didChange bool, err error) {
	if !sku.TransactedLessor.LessPtr(&store.config.Sku, kt1) {
		return
	}

	store.config.lock.Lock()
	defer store.config.lock.Unlock()

	didChange = true

	sku.Resetter.ResetWith(&store.config.Sku, kt1)

	store.config.setNeedsRecompile(
		fmt.Sprintf("updated konfig: %s", &store.config.Sku),
	)

	if err = store.loadMutableConfigBlob(
		store.config.Sku.GetType(),
		store.config.Sku.GetBlobSha(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
