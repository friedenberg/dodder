package organize_text

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/echo/format"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type writer struct {
	sku.ObjectFactory
	OmitLeadingEmptyLine bool
	object_metadata.Metadata
	*format.LineWriter
	maxDepth int
	options  Options
}

func (av writer) write(a *Assignment) (err error) {
	spaceCount := av.maxDepth

	hinMaxWidth := 3

	if spaceCount < hinMaxWidth {
		spaceCount = hinMaxWidth
	}

	tab_prefix := strings.Repeat(" ", spaceCount+1)

	if a.GetDepth() == 0 && !av.OmitLeadingEmptyLine {
		av.WriteExactlyOneEmpty()
	} else if a.GetDepth() < 0 {
		err = errors.ErrorWithStackf("negative depth: %d", a.GetDepth())
		return err
	}

	if a.Transacted.Metadata.Tags != nil && a.Transacted.Metadata.Tags.Len() > 0 {
		sharps := strings.Repeat("#", a.GetDepth())
		alignmentSpacing := strings.Repeat(" ", a.AlignmentSpacing())

		av.WriteLines(
			fmt.Sprintf(
				"%s%s %s%s",
				tab_prefix[len(sharps)-1:],
				sharps,
				alignmentSpacing,
				quiter.StringCommaSeparated(a.Transacted.Metadata.Tags),
			),
		)
		av.WriteExactlyOneEmpty()
	}

	write := func(object *obj) (err error) {
		var sb strings.Builder

		if object.tipe.IsDirectOrSelf() {
			sb.WriteString("- ")
		} else {
			sb.WriteString("% ")
		}

		cursor := object.sku.Clone()
		cursorExternal := cursor.GetSkuExternal()
		cursorExternal.Metadata.Subtract(&av.Metadata)
		mes := cursorExternal.GetMetadata().GetTags().CloneMutableSetPtrLike()

		if err = a.SubtractFromSet(mes); err != nil {
			err = errors.Wrap(err)
			return err
		}

		cursorExternal.Metadata.SetTags(mes)

		if _, err = av.options.fmtBox.EncodeStringTo(
			cursor,
			&sb,
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

		av.WriteStringers(&sb)

		return err
	}

	a.Objects.Sort()

	for _, z := range a.Objects {
		if err = write(z); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	if a.Objects.Len() > 0 {
		av.WriteExactlyOneEmpty()
	}

	for _, c := range a.Children {
		av.write(c)
	}

	av.WriteExactlyOneEmpty()

	return err
}
