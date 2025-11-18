package box_format

import (
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

type CheckedOutHeaderString string

func (f CheckedOutHeaderString) WriteBoxHeader(
	header *string_format_writer.BoxHeader,
	co *sku.CheckedOut,
) (err error) {
	header.RightAligned = true
	header.Value = string(f)

	return err
}
