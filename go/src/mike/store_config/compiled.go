package store_config

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/comments"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type (
	configRepo    = repo_configs.Config
	configGenesis = genesis_configs.ConfigPrivate
	CLI           = repo_config_cli.Blob

	Config struct {
		*compiled

		configRepo
		configGenesis
		CLI
	}

	compiled struct {
		// TODO move to store
		lock sync.Mutex

		// TODO move to store
		changes []string

		// TODO move to store
		Sku sku.Transacted

		Tags         interfaces.MutableSetLike[*tag]
		ImplicitTags implicitTagMap

		// Typen
		ExtensionsToTypes map[string]string
		TypesToExtensions map[string]string
		Types             sku.TransactedMutableSet
		InlineTypes       interfaces.SetLike[values.String]

		// Kasten
		Repos sku.TransactedMutableSet
	}
)

func (k *compiled) GetSku() *sku.Transacted {
	return &k.Sku
}

func (k *compiled) addRepo(
	c *sku.Transacted,
) (didChange bool, err error) {
	k.lock.Lock()
	defer k.lock.Unlock()

	b := sku.GetTransactedPool().Get()

	sku.Resetter.ResetWith(b, c)

	if didChange, err = quiter.AddOrReplaceIfGreater(
		k.Repos,
		b,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (k *compiled) addType(
	b1 *sku.Transacted,
) (didChange bool, err error) {
	if err = genres.Type.AssertGenre(b1); err != nil {
		err = errors.Wrap(err)
		return
	}

	b := sku.GetTransactedPool().Get()

	sku.Resetter.ResetWith(b, b1)

	k.lock.Lock()
	defer k.lock.Unlock()

	if didChange, err = quiter.AddOrReplaceIfGreater(
		k.Types,
		b,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (config *Config) GetTypeStringFromExtension(t string) string {
	return config.ExtensionsToTypes[t]
}

func (config *Config) GetTypeExtension(v string) string {
	return config.TypesToExtensions[v]
}

func (config Config) IsInlineType(tipe ids.Type) (isInline bool) {
	comments.Change("fix this horrible hack")
	if tipe.IsEmpty() {
		return true
	}

	isInline = config.InlineTypes.ContainsKey(tipe.String()) ||
		ids.IsBuiltin(tipe)

	return
}
