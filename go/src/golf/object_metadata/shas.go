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

// TODO make this a map
type Shas struct {
	Blob                         sha.Sha
	SelfMetadata                 sha.Sha
	SelfMetadataWithoutTai       sha.Sha
	SelfMetadataObjectIdParent   sha.Sha
	ParentMetadataObjectIdParent sha.Sha
}

func (s *Shas) Reset() {
	s.Blob.Reset()
	s.SelfMetadata.Reset()
	s.SelfMetadataWithoutTai.Reset()
	s.SelfMetadataObjectIdParent.Reset()
	s.ParentMetadataObjectIdParent.Reset()
}

func (dst *Shas) ResetWith(src *Shas) {
	dst.Blob.ResetWith(&src.Blob)
	dst.SelfMetadata.ResetWith(&src.SelfMetadata)
	dst.SelfMetadataWithoutTai.ResetWith(&src.SelfMetadataWithoutTai)
	dst.SelfMetadataObjectIdParent.ResetWith(&src.SelfMetadataObjectIdParent)
	dst.ParentMetadataObjectIdParent.ResetWith(&src.ParentMetadataObjectIdParent)
}

func (s *Shas) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%s: %s\n", "Blob", &s.Blob)
	fmt.Fprintf(&sb, "%s: %s\n", ShaKeySelfMetadata, &s.SelfMetadata)
	fmt.Fprintf(&sb, "%s: %s\n", ShaKeySelfMetadataWithouTai, &s.SelfMetadataWithoutTai)

	return sb.String()
}

func (s *Shas) Add(k, v string) (err error) {
	switch k {
	case ShaKeySelfMetadata:
		if err = s.SelfMetadata.Set(v); err != nil {
			err = errors.Wrap(err)
			return
		}

	case ShaKeySelfMetadataWithouTai:
		if err = s.SelfMetadataWithoutTai.Set(v); err != nil {
			err = errors.Wrap(err)
			return
		}

	case ShaKeySelfMetadataObjectIdParent:
		if err = s.SelfMetadataObjectIdParent.Set(v); err != nil {
			err = errors.Wrap(err)
			return
		}

	case ShaKeyParentMetadataObjectIdParent:
		if err = s.ParentMetadataObjectIdParent.Set(v); err != nil {
			err = errors.Wrap(err)
			return
		}

	default:
		err = errors.ErrorWithStackf("unrecognized sha kind: %q", k)
		return
	}

	return
}
