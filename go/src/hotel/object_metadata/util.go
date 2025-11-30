package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/expansion"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

func ExpandTags(
	metadata IMetadata,
	expander expansion.Expander,
) interfaces.Seq[ids.Tag] {
	ids.ExpandMany(
		metadata.GetTags().All(),
		expander,
	)

	return nil
}
