package object_metadata

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

// TODO consider moving all non-essential digests to a separate key-value store
type Digests struct {
	BlobId      sha.Sha
	FingerPrint sha.Sha

	// TODO move cache
	SelfMetadataWithoutTai       sha.Sha
	ParentMetadataObjectIdParent sha.Sha
}

func (digests *Digests) Reset() {
	digests.BlobId.Reset()
	digests.SelfMetadataWithoutTai.Reset()
	digests.FingerPrint.Reset()
	digests.ParentMetadataObjectIdParent.Reset()
}

func (digests *Digests) ResetWith(src *Digests) {
	digests.BlobId.ResetWith(&src.BlobId)
	digests.FingerPrint.ResetWith(
		&src.FingerPrint,
	)
	digests.ParentMetadataObjectIdParent.ResetWith(
		&src.ParentMetadataObjectIdParent,
	)
}

func (digests *Digests) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%s: %s\n", "Blob", &digests.BlobId)

	return sb.String()
}
