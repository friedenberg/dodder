package object_metadata

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/keys"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

const (
	ShaKeySelfMetadataObjectIdParent   = keys.ShaKeySelfMetadataObjectIdParent
	ShaKeyParentMetadataObjectIdParent = keys.ShaKeyParentMetadataObjectIdParent
	ShaKeySelf                         = keys.ShaKeySelf
	ShaKeyParent                       = keys.ShaKeyParent
)

type Sha struct {
	*sha.Sha
	string
}

// TODO consider moving all non-essential digests to a separate key-value store
type Shas struct {
	Blob                         sha.Sha
	SelfMetadataWithoutTai       sha.Sha
	SelfMetadataObjectIdParent   sha.Sha
	ParentMetadataObjectIdParent sha.Sha
}

func (shas *Shas) Reset() {
	shas.Blob.Reset()
	shas.SelfMetadataWithoutTai.Reset()
	shas.SelfMetadataObjectIdParent.Reset()
	shas.ParentMetadataObjectIdParent.Reset()
}

func (shas *Shas) ResetWith(src *Shas) {
	shas.Blob.ResetWith(&src.Blob)
	shas.SelfMetadataObjectIdParent.ResetWith(&src.SelfMetadataObjectIdParent)
	shas.ParentMetadataObjectIdParent.ResetWith(
		&src.ParentMetadataObjectIdParent,
	)
}

func (shas *Shas) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%s: %s\n", "Blob", &shas.Blob)

	return sb.String()
}

func (shas *Shas) Add(k, v string) (err error) {
	switch k {
	case ShaKeySelfMetadataObjectIdParent:
		if err = shas.SelfMetadataObjectIdParent.Set(v); err != nil {
			err = errors.Wrap(err)
			return
		}

	case ShaKeyParentMetadataObjectIdParent:
		if err = shas.ParentMetadataObjectIdParent.Set(v); err != nil {
			err = errors.Wrap(err)
			return
		}

	default:
		err = errors.ErrorWithStackf("unrecognized sha kind: %q", k)
		return
	}

	return
}
