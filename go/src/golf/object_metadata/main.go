package object_metadata

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/flag_policy"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/flag"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/descriptions"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type Field = string_format_writer.Field

type Metadata struct {
	// Domain
	RepoPubKey repo_signing.PublicKey
	RepoSig    repo_signing.Data
	// InventoryListTai

	Description descriptions.Description
	Tags        ids.TagMutableSet // public for gob, but should be private
	Type        ids.Type

	Shas
	Tai ids.Tai

	Comments []string
	Cache    Cache
	Fields   []Field
}

func (m *Metadata) GetMetadata() *Metadata {
	return m
}

func (m *Metadata) Sha() *sha.Sha {
	return &m.SelfMetadataObjectIdParent
}

func (m *Metadata) Mutter() *sha.Sha {
	return &m.ParentMetadataObjectIdParent
}

// TODO replace with command_components.ObjectMetadata
func (m *Metadata) SetFlagSet(f *flag.FlagSet) {
	m.SetFlagSetDescription(
		f,
		"the description to use for created or updated Zettels",
	)

	m.SetFlagSetTags(
		f,
		"the tags to use for created or updated object",
	)

	m.SetFlagSetType(
		f,
		"the type for the created or updated object",
	)
}

func (m *Metadata) SetFlagSetDescription(f *flag.FlagSet, usage string) {
	f.Var(
		&m.Description,
		"description",
		usage,
	)
}

func (m *Metadata) SetFlagSetTags(f *flag.FlagSet, usage string) {
	// TODO add support for tag_paths
	fes := flag.Make(
		flag_policy.FlagPolicyAppend,
		func() string {
			return m.Cache.TagPaths.String()
		},
		func(value string) (err error) {
			values := strings.Split(value, ",")

			for _, tagString := range values {
				if err = m.AddTagString(tagString); err != nil {
					err = errors.Wrap(err)
					return
				}
			}

			return
		},
		func() {
			m.ResetTags()
		},
	)

	f.Var(
		fes,
		"tags",
		usage,
	)
}

func (m *Metadata) SetFlagSetType(f *flag.FlagSet, usage string) {
	f.Func(
		"type",
		usage,
		func(v string) (err error) {
			return m.Type.Set(v)
		},
	)
}

func (z *Metadata) UserInputIsEmpty() bool {
	if !z.Description.IsEmpty() {
		return false
	}

	if z.Tags != nil && z.Tags.Len() > 0 {
		return false
	}

	if !ids.IsEmpty(z.Type) {
		return false
	}

	return true
}

func (z *Metadata) IsEmpty() bool {
	if !z.Blob.IsNull() {
		return false
	}

	if !z.UserInputIsEmpty() {
		return false
	}

	if !z.Tai.IsZero() {
		return false
	}

	return true
}

// TODO fix issue with GetTags being nil sometimes
func (m *Metadata) GetTags() ids.TagSet {
	if m.Tags == nil {
		m.Tags = ids.MakeTagMutableSet()
	}

	return m.Tags
}

func (m *Metadata) ResetTags() {
	if m.Tags == nil {
		m.Tags = ids.MakeTagMutableSet()
	}

	m.Tags.Reset()
	m.Cache.TagPaths.Reset()
}

func (z *Metadata) AddTagString(es string) (err error) {
	if es == "" {
		return
	}

	var e ids.Tag

	if err = e.Set(es); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = z.AddTagPtr(&e); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (m *Metadata) AddTagPtr(e *ids.Tag) (err error) {
	if e == nil || e.String() == "" {
		return
	}

	if m.Tags == nil {
		m.Tags = ids.MakeTagMutableSet()
	}

	ids.AddNormalizedTag(m.Tags, e)
	cs := catgut.MakeFromString(e.String())
	m.Cache.TagPaths.AddTag(cs)

	return
}

func (m *Metadata) AddTagPtrFast(e *ids.Tag) (err error) {
	if m.Tags == nil {
		m.Tags = ids.MakeTagMutableSet()
	}

	if err = m.Tags.Add(*e); err != nil {
		err = errors.Wrap(err)
		return
	}

	cs := catgut.MakeFromString(e.String())

	if err = m.Cache.TagPaths.AddTag(cs); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (metadata *Metadata) SetTags(tags ids.TagSet) {
	if metadata.Tags == nil {
		metadata.Tags = ids.MakeTagMutableSet()
	}

	metadata.Tags.Reset()

	if tags == nil {
		return
	}

	if tags.Len() == 1 && tags.Any().String() == "" {
		panic("empty tag set")
	}

	errors.PanicIfError(tags.EachPtr(metadata.AddTagPtr))
}

func (z *Metadata) GetType() ids.Type {
	return z.Type
}

func (z *Metadata) GetTypePtr() *ids.Type {
	return &z.Type
}

func (z *Metadata) GetTai() ids.Tai {
	return z.Tai
}

// TODO-P2 remove
func (b *Metadata) EqualsSansTai(a *Metadata) bool {
	return EqualerSansTai.Equals(a, b)
}

// TODO-P2 remove
func (pz *Metadata) Equals(z1 *Metadata) bool {
	return Equaler.Equals(pz, z1)
}

func (a *Metadata) Subtract(
	b *Metadata,
) {
	if a.Type.String() == b.Type.String() {
		a.Type = ids.Type{}
	}

	if a.Tags == nil {
		return
	}

	// ui.Debug().Print("before", b.Tags, a.Tags)

	for e := range b.Tags.AllPtr() {
		// ui.Debug().Print(e)
		a.Tags.DelPtr(e)
	}

	// ui.Debug().Print("after", b.Tags, a.Tags)
}

func (mp *Metadata) AddComment(f string, vals ...interface{}) {
	mp.Comments = append(mp.Comments, fmt.Sprintf(f, vals...))
}

func (selbst *Metadata) SetMutter(mg Getter) (err error) {
	mutter := mg.GetMetadata()

	if err = selbst.Mutter().SetShaLike(
		mutter.Sha(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = selbst.ParentMetadataObjectIdParent.SetShaLike(
		&mutter.SelfMetadataObjectIdParent,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (m *Metadata) GenerateExpandedTags() {
	m.Cache.SetExpandedTags(ids.ExpandMany(
		m.GetTags(),
		expansion.ExpanderRight,
	))
}
