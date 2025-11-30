package objects

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/foxtrot/descriptions"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type builder struct {
	metadata *metadata
}

func MakeBuilder() *builder {
	return &builder{
		metadata: &metadata{},
	}
}

func (builder *builder) checkReuse() {
	if builder.metadata == nil {
		panic("attempting to use consumed builder")
	}
}

func (builder *builder) WithType(typeString string) *builder {
	builder.checkReuse()
	errors.PanicIfError(builder.metadata.GetTypeMutable().Set(typeString))
	return builder
}

func (builder *builder) WithDescription(
	descriptionString string,
) *builder {
	builder.checkReuse()
	builder.metadata.Description.ResetWith(descriptions.Make(descriptionString))
	return builder
}

func (builder *builder) WithTags(tags ids.TagSet) *builder {
	builder.checkReuse()
	builder.metadata.SetTags(tags)
	return builder
}

func (builder *builder) WithBlobDigest(digest markl.Id) *builder {
	builder.checkReuse()
	builder.metadata.GetBlobDigestMutable().ResetWithMarklId(digest)
	return builder
}

func (builder *builder) Build() metadata {
	metadata := *builder.metadata
	builder.metadata = nil
	return metadata
}
