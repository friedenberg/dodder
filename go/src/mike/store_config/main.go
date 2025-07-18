package store_config

import (
	"encoding/gob"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/comments"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
	"code.linenisgreat.com/dodder/go/src/golf/repo_config_blobs"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
)

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
	immutable_config_private = genesis_configs.Private
	cli                      = repo_config_cli.Config
	ApproximatedType         = typed_blob_store.ApproximatedType

	Store interface {
		interfaces.Config
		genesis_configs.Private

		repo_config_blobs.Getter

		ids.InlineTypeChecker
		GetTypeExtension(string) string
		GetCLIConfig() repo_config_cli.Config
		GetImmutableConfig() genesis_configs.Private
		GetFileExtensions() interfaces.FileExtensions
		HasChanges() (ok bool)
		GetChanges() (out []string)

		GetTagOrRepoIdOrType(
			v string,
		) (sk *sku.Transacted, err error)
		GetImplicitTags(*ids.Tag) ids.TagSet
		GetApproximatedType(
			k interfaces.ObjectId,
		) (ct ApproximatedType)
		GetSku() *sku.Transacted
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
	cli
	compiled
	mutable_config_blob
	immutable_config_private
}

func (store *store) GetCLIConfig() repo_config_cli.Config {
	return store.cli
}

func (store *store) Reset() error {
	store.Blob = repo_config_blobs.V1{}
	store.ExtensionsToTypes = make(map[string]string)
	store.TypesToExtensions = make(map[string]string)

	store.Tags = collections_value.MakeMutableValueSet[*tag](nil)
	store.InlineTypes = collections_value.MakeMutableValueSet[values.String](
		nil,
	)
	store.ImplicitTags = make(implicitTagMap)
	store.Repos = sku.MakeTransactedMutableSet()
	store.Types = sku.MakeTransactedMutableSet()

	sku.TransactedResetter.Reset(&store.Sku)

	return nil
}

func (store *store) GetMutableConfig() repo_config_blobs.Blob {
	return store.mutable_config_blob
}

func (store *store) Initialize(
	envRepo env_repo.Env,
	kcli repo_config_cli.Config,
) (err error) {
	store.cli = kcli
	store.Reset()
	store.immutable_config_private = envRepo.GetConfigPrivate().Blob

	store.typedConfigBlobStore = typed_blob_store.MakeConfigStore(envRepo)

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

	store.cli.ApplyPrintOptionsConfig(store.GetPrintOptions())

	return
}

func (store *store) SetCli(k repo_config_cli.Config) {
	store.cli = k
}

func (store *store) SetCliFromCommander(k repo_config_cli.Config) {
	oldBasePath := store.BasePath
	store.cli = k
	store.BasePath = oldBasePath
}

func (store *store) IsDryRun() bool {
	return store.cli.IsDryRun()
}

func (store *store) SetDryRun(v bool) {
	store.cli.SetDryRun(v)
}

func (store *store) GetTypeStringFromExtension(t string) string {
	return store.ExtensionsToTypes[t]
}

func (store *store) GetTypeExtension(v string) string {
	return store.TypesToExtensions[v]
}

func (store *store) AddTransacted(
	child *sku.Transacted,
	parent *sku.Transacted,
) (err error) {
	didChange := false

	g := child.ObjectId.GetGenre()

	switch g {
	case genres.Type:
		if didChange, err = store.addType(child); err != nil {
			err = errors.Wrap(err)
			return
		}

		if didChange {
			store.SetNeedsRecompile(fmt.Sprintf("modified type: %s", child))
		}

		return

	case genres.Tag:
		if didChange, err = store.addTag(child, parent); err != nil {
			err = errors.Wrap(err)
			return
		}

		var tag ids.Tag

		if err = tag.TodoSetFromObjectId(child.GetObjectId()); err != nil {
			err = errors.Wrap(err)
			return
		}

		if child.Metadata.GetTags().Len() > 0 {
			store.SetNeedsRecompile(
				fmt.Sprintf(
					"tag with tags added: %q -> %q",
					tag,
					quiter.SortedValues(child.Metadata.GetTags()),
				),
			)
		}

	case genres.Repo:
		if didChange, err = store.addRepo(child); err != nil {
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

	store.SetNeedsRecompile(fmt.Sprintf("modified: %s", child))

	return
}

func (store *store) IsInlineType(k ids.Type) (isInline bool) {
	comments.Change("fix this horrible hack")
	if k.IsEmpty() {
		return true
	}

	isInline = store.InlineTypes.ContainsKey(k.String()) ||
		ids.IsBuiltin(k)

	return
}
