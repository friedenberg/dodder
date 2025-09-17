package command_components

import (
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
	"code.linenisgreat.com/dodder/go/src/kilo/inventory_list_coders"
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
