package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/foxtrot/descriptions"
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
	errors.PanicIfError(builder.metadata.Type.Set(typeString))
	return builder
}

func (builder *builder) WithDescription(
	descriptionString string,
) *builder {
	builder.checkReuse()
	builder.metadata.Description.ResetWith(descriptions.Make(descriptionString))
	return builder
}

func (builder *builder) Build() IMetadataMutable {
	metadata := builder.metadata
	builder.metadata = nil
	return metadata
}
