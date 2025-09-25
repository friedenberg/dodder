package box_format

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type CheckedOutHeaderDeleted struct {
	interfaces.ConfigDryRunGetter
}

func (f CheckedOutHeaderDeleted) WriteBoxHeader(
	header *string_format_writer.BoxHeader,
	co *sku.CheckedOut,
) (err error) {
	header.RightAligned = true

	if f.IsDryRun() {
		header.Value = "would delete"
	} else {
		header.Value = "deleted"
	}

	return err
}
