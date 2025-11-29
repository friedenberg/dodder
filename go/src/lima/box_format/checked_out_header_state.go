package box_format

import (
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/hotel/object_metadata"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

type CheckedOutHeaderState struct{}

func (f CheckedOutHeaderState) WriteBoxHeader(
	header *string_format_writer.BoxHeader,
	checkedOut *sku.CheckedOut,
) (err error) {
	header.RightAligned = true

	state := checkedOut.GetState()
	stateString := state.String()

	switch state {
	case checked_out_state.CheckedOut:
		if object_metadata.EqualerSansTai.Equals(
			checkedOut.GetSku().GetMetadata(),
			checkedOut.GetSkuExternal().GetSku().GetMetadata(),
		) {
			header.Value = string_format_writer.StringSame
		} else {
			header.Value = string_format_writer.StringChanged
		}

	default:
		header.Value = stateString
	}

	return err
}
