package object_metadata

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

type Digests struct {
	// TODO transform into byte slices
	Blob   sha.Sha
	Self   sha.Sha
	Mother sha.Sha

	// TODO moving to a separate key-value store
	SelfWithoutTai sha.Sha
}

func (digests *Digests) GetDigest() *sha.Sha {
	return &digests.Self
}

func (digests *Digests) GetMotherDigest() *sha.Sha {
	return &digests.Mother
}

func (digests *Digests) Reset() {
	digests.Blob.Reset()
	digests.SelfWithoutTai.Reset()
	digests.Self.Reset()
	digests.Mother.Reset()
}

func (digests *Digests) ResetWith(src *Digests) {
	digests.Blob.ResetWith(&src.Blob)
	digests.Self.ResetWith(
		&src.Self,
	)
	digests.Mother.ResetWith(
		&src.Mother,
	)
}

func (digests *Digests) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%s: %s\n", "Blob", &digests.Blob)

	return sb.String()
}
