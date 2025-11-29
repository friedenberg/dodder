package organize_text

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter_set"
	"code.linenisgreat.com/dodder/go/src/delta/ohio"
	"code.linenisgreat.com/dodder/go/src/echo/format"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/golf/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/hotel/object_metadata"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

type TagSetGetter interface {
	GetTags() ids.TagSet
}

func NewMetadata(repoId ids.RepoId) Metadata {
	return Metadata{
		RepoId:           repoId,
		TagSet:           ids.MakeTagSetFromSlice(),
		OptionCommentSet: MakeOptionCommentSet(nil),
	}
}

func NewMetadataWithOptionCommentLookup(
	repoId ids.RepoId,
	elements map[string]OptionComment,
) Metadata {
	return Metadata{
		RepoId:           repoId,
		TagSet:           ids.MakeTagSetFromSlice(),
		OptionCommentSet: MakeOptionCommentSet(elements),
	}
}

// TODO replace with embedded *sku.Transacted
type Metadata struct {
	ids.TagSet
	Matchers interfaces.Set[sku.Query] // TODO remove
	OptionCommentSet
	Type   ids.Type
	RepoId ids.RepoId
}

func (metadata *Metadata) GetTags() ids.TagSet {
	return metadata.TagSet
}

func (metadata *Metadata) SetFromObjectMetadata(
	otherMetadata object_metadata.IMetadataMutable,
	repoId ids.RepoId,
) (err error) {
	metadata.TagSet = ids.CloneTagSet(otherMetadata.GetTags())

	for comment := range otherMetadata.GetIndex().GetComments() {
		if err = metadata.OptionCommentSet.Set(comment); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	metadata.Type = otherMetadata.GetType()

	return err
}

func (metadata Metadata) RemoveFromTransacted(object sku.SkuType) (err error) {
	tags := ids.CloneTagSetMutable(object.GetSkuExternal().GetMetadata().GetTags())

	for element := range metadata.All() {
		quiter_set.Del(tags, element)
	}

	object.GetSkuExternal().GetMetadataMutable().SetTags(tags)

	return err
}

func (metadata Metadata) AsMetadata() (m1 object_metadata.IMetadataMutable) {
	m1 = object_metadata.Make()
	m1.GetTypeMutable().ResetWith(metadata.Type)
	m1.SetTags(metadata.TagSet)
	return m1
}

func (metadata Metadata) GetMetadataWriterTo() triple_hyphen_io.MetadataWriterTo {
	return metadata
}

func (metadata Metadata) HasMetadataContent() bool {
	if metadata.Len() > 0 {
		return true
	}

	if !metadata.Type.IsEmpty() {
		return true
	}

	if len(metadata.OptionCommentSet.OptionComments) > 0 {
		return true
	}

	return false
}

func (metadata *Metadata) ReadFrom(reader io.Reader) (n int64, err error) {
	bufferedReader, repool := pool.GetBufferedReader(reader)
	defer repool()

	tagSet := ids.MakeTagSetMutable()

	if n, err = format.ReadLines(
		bufferedReader,
		ohio.MakeLineReaderRepeat(
			ohio.MakeLineReaderKeyValues(
				map[string]interfaces.FuncSetString{
					"%": metadata.OptionCommentSet.Set,
					"-": quiter.MakeFuncSetString(tagSet),
					"!": metadata.Type.Set,
				},
			),
		),
	); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	metadata.TagSet = ids.CloneTagSet(tagSet)

	return n, err
}

func (metadata Metadata) WriteTo(w1 io.Writer) (n int64, err error) {
	w := format.NewLineWriter()

	for _, o := range metadata.OptionCommentSet.OptionComments {
		w.WriteFormat("%% %s", o)
	}

	for _, e := range quiter.SortedStrings(metadata.TagSet) {
		w.WriteFormat("- %s", e)
	}

	tString := metadata.Type.StringSansOp()

	if tString != "" {
		w.WriteFormat("! %s", tString)
	}

	if metadata.Matchers != nil {
		for _, c := range quiter.SortedStrings(metadata.Matchers) {
			w.WriteFormat("%% Matcher:%s", c)
		}
	}

	return w.WriteTo(w1)
}
