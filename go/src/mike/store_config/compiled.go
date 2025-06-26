package store_config

import (
	"fmt"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type compiled struct {
	lock sync.Mutex

	changes []string

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

func (k *compiled) GetSku() *sku.Transacted {
	return &k.Sku
}

func (k *store) setTransacted(
	kt1 *sku.Transacted,
) (didChange bool, err error) {
	if !sku.TransactedLessor.LessPtr(&k.Sku, kt1) {
		return
	}

	k.lock.Lock()
	defer k.lock.Unlock()

	didChange = true

	sku.Resetter.ResetWith(&k.Sku, kt1)

	k.setNeedsRecompile(fmt.Sprintf("updated konfig: %s", &k.Sku))

	if err = k.loadMutableConfigBlob(
		k.Sku.GetType(),
		k.Sku.GetBlobSha(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
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
