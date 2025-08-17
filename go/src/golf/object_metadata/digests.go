package object_metadata

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

// TODO consider moving all non-essential digests to a separate key-value store
type Digests struct {
	Blob                         sha.Sha
	SelfMetadataWithoutTai       sha.Sha
	SelfMetadataObjectIdParent   sha.Sha
	ParentMetadataObjectIdParent sha.Sha
}

func (digests *Digests) Reset() {
	digests.Blob.Reset()
	digests.SelfMetadataWithoutTai.Reset()
	digests.SelfMetadataObjectIdParent.Reset()
	digests.ParentMetadataObjectIdParent.Reset()
}

func (digests *Digests) ResetWith(src *Digests) {
	digests.Blob.ResetWith(&src.Blob)
	digests.SelfMetadataObjectIdParent.ResetWith(
		&src.SelfMetadataObjectIdParent,
	)
	digests.ParentMetadataObjectIdParent.ResetWith(
		&src.ParentMetadataObjectIdParent,
	)
}

func (digests *Digests) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%s: %s\n", "Blob", &digests.Blob)

	return sb.String()
}
