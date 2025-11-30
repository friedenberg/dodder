package objects

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

func ExpandTags(
	metadata Metadata,
	expander expansion.Expander,
) interfaces.Seq[ids.Tag] {
	expansion.ExpandMany(
		metadata.GetTags().All(),
		expander,
	)

	return nil
}
