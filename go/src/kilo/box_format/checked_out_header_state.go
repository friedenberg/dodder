package box_format

import (
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type CheckedOutHeaderState struct{}

func (f CheckedOutHeaderState) WriteBoxHeader(
	header *string_format_writer.BoxHeader,
	co *sku.CheckedOut,
) (err error) {
	header.RightAligned = true

	state := co.GetState()
	stateString := state.String()

	switch state {
	case checked_out_state.CheckedOut:
		if co.GetSku().Metadata.EqualsSansTai(&co.GetSkuExternal().GetSku().Metadata) {
			header.Value = string_format_writer.StringSame
		} else {
			header.Value = string_format_writer.StringChanged
		}

	default:
		header.Value = stateString
	}

	return err
}
