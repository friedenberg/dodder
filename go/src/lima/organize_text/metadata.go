package organize_text

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/echo/format"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io2"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type TagSetGetter interface {
	GetTags() ids.TagSet
}

func NewMetadata(repoId ids.RepoId) Metadata {
	return Metadata{
		RepoId:           repoId,
		TagSet:           ids.MakeTagSet(),
		OptionCommentSet: MakeOptionCommentSet(nil),
	}
}

func NewMetadataWithOptionCommentLookup(
	repoId ids.RepoId,
	elements map[string]OptionComment,
) Metadata {
	return Metadata{
		RepoId:           repoId,
		TagSet:           ids.MakeTagSet(),
		OptionCommentSet: MakeOptionCommentSet(elements),
	}
}

// TODO replace with embedded *sku.Transacted
type Metadata struct {
	ids.TagSet
	Matchers interfaces.SetLike[sku.Query] // TODO remove
	OptionCommentSet
	Type   ids.Type
	RepoId ids.RepoId
}

func (m *Metadata) GetTags() ids.TagSet {
	return m.TagSet
}

func (m *Metadata) SetFromObjectMetadata(
	om *object_metadata.Metadata,
	repoId ids.RepoId,
) (err error) {
	m.TagSet = om.Tags.CloneSetPtrLike()

	for _, c := range om.Comments {
		if err = m.OptionCommentSet.Set(c); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	m.Type = om.Type

	return
}

func (m Metadata) RemoveFromTransacted(sk sku.SkuType) (err error) {
	mes := sk.GetSkuExternal().Metadata.GetTags().CloneMutableSetPtrLike()

	if err = m.Each(mes.Del); err != nil {
		err = errors.Wrap(err)
		return
	}

	sk.GetSkuExternal().Metadata.SetTags(mes)

	return
}

func (m Metadata) AsMetadata() (m1 object_metadata.Metadata) {
	m1.Type = m.Type
	m1.SetTags(m.TagSet)
	return
}

func (m Metadata) GetMetadataWriterTo() triple_hyphen_io2.MetadataWriterTo {
	return m
}

func (m Metadata) HasMetadataContent() bool {
	if m.Len() > 0 {
		return true
	}

	if !m.Type.IsEmpty() {
		return true
	}

	if len(m.OptionCommentSet.OptionComments) > 0 {
		return true
	}

	return false
}

func (m *Metadata) ReadFrom(r1 io.Reader) (n int64, err error) {
	r := bufio.NewReader(r1)

	mes := ids.MakeTagMutableSet()

	if n, err = format.ReadLines(
		r,
		ohio.MakeLineReaderRepeat(
			ohio.MakeLineReaderKeyValues(
				map[string]interfaces.FuncSetString{
					"%": m.OptionCommentSet.Set,
					"-": quiter.MakeFuncSetString(mes),
					"!": m.Type.Set,
				},
			),
		),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	m.TagSet = mes.CloneSetPtrLike()

	return
}

func (m Metadata) WriteTo(w1 io.Writer) (n int64, err error) {
	w := format.NewLineWriter()

	for _, o := range m.OptionCommentSet.OptionComments {
		w.WriteFormat("%% %s", o)
	}

	for _, e := range quiter.SortedStrings(m.TagSet) {
		w.WriteFormat("- %s", e)
	}

	tString := m.Type.StringSansOp()

	if tString != "" {
		w.WriteFormat("! %s", tString)
	}

	if m.Matchers != nil {
		for _, c := range quiter.SortedStrings(m.Matchers) {
			w.WriteFormat("%% Matcher:%s", c)
		}
	}

	return w.WriteTo(w1)
}
