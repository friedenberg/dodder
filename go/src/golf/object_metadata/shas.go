package object_metadata

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/keys"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

const (
	ShaKeySelfMetadata                 = keys.ShaKeySelfMetadata
	ShaKeySelfMetadataWithouTai        = keys.ShaKeySelfMetadataWithouTai
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
	SelfMetadata                 sha.Sha
	SelfMetadataWithoutTai       sha.Sha
	SelfMetadataObjectIdParent   sha.Sha
	ParentMetadataObjectIdParent sha.Sha
}

func (shas *Shas) Reset() {
	shas.Blob.Reset()
	shas.SelfMetadata.Reset()
	shas.SelfMetadataWithoutTai.Reset()
	shas.SelfMetadataObjectIdParent.Reset()
	shas.ParentMetadataObjectIdParent.Reset()
}

func (shas *Shas) ResetWith(src *Shas) {
	shas.Blob.ResetWith(&src.Blob)
	shas.SelfMetadata.ResetWith(&src.SelfMetadata)
	shas.SelfMetadataWithoutTai.ResetWith(&src.SelfMetadataWithoutTai)
	shas.SelfMetadataObjectIdParent.ResetWith(&src.SelfMetadataObjectIdParent)
	shas.ParentMetadataObjectIdParent.ResetWith(
		&src.ParentMetadataObjectIdParent,
	)
}

func (shas *Shas) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%s: %s\n", "Blob", &shas.Blob)
	fmt.Fprintf(&sb, "%s: %s\n", ShaKeySelfMetadata, &shas.SelfMetadata)
	fmt.Fprintf(
		&sb,
		"%s: %s\n",
		ShaKeySelfMetadataWithouTai,
		&shas.SelfMetadataWithoutTai,
	)

	return sb.String()
}

func (shas *Shas) Add(k, v string) (err error) {
	switch k {
	case ShaKeySelfMetadata:
		if err = shas.SelfMetadata.Set(v); err != nil {
			err = errors.Wrap(err)
			return
		}

	case ShaKeySelfMetadataWithouTai:
		if err = shas.SelfMetadataWithoutTai.Set(v); err != nil {
			err = errors.Wrap(err)
			return
		}

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
