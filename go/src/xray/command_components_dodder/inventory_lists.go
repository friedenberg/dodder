package command_components_dodder

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/kilo/env_repo"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
	"code.linenisgreat.com/dodder/go/src/lima/box_format"
	"code.linenisgreat.com/dodder/go/src/mike/inventory_list_coders"
)

type InventoryLists struct{}

func (InventoryLists) MakeInventoryListCoderCloset(
	envRepo env_repo.Env,
) inventory_list_coders.Closet {
	boxFormat := box_format.MakeBoxTransactedArchive(
		envRepo,
		options_print.Options{}.WithPrintTai(true),
	)

	return inventory_list_coders.MakeCloset(
		envRepo,
		boxFormat,
	)
}

func (InventoryLists) MakeSeqFromPath(
	ctx interfaces.ActiveContext,
	inventoryListCoderCloset inventory_list_coders.Closet,
	inventoryListPath string,
	afterDecoding func(*sku.Transacted) error,
) interfaces.SeqError[*sku.Transacted] {
	var readCloser io.ReadCloser

	// setup inventory list reader
	{
		var err error

		if readCloser, err = files.Open(
			inventoryListPath,
		); err != nil {
			ctx.Cancel(err)
			return nil
		}
	}

	bufferedReader, repoolBufferedReader := pool.GetBufferedReader(readCloser)

	seq := inventoryListCoderCloset.AllDecodedObjectsFromStream(
		bufferedReader,
		afterDecoding,
	)

	return func(yield func(*sku.Transacted, error) bool) {
		defer errors.ContextMustClose(ctx, readCloser)
		defer repoolBufferedReader()

		for object, err := range seq {
			if !yield(object, err) {
				return
			}
		}
	}
}
