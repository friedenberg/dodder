package sku

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/comments"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/golf/fd"
	"code.linenisgreat.com/dodder/go/src/hotel/object_metadata"
	"code.linenisgreat.com/dodder/go/src/hotel/repo_configs"
)

func MakeProto(defaults repo_configs.Defaults) (proto Proto) {
	var tipe ids.Type
	var tags ids.TagSet

	if defaults != nil {
		tipe = defaults.GetDefaultType()
		tags = ids.MakeTagSetFromSlice(defaults.GetDefaultTags()...)
	}

	proto.Metadata.GetTypeMutable().ResetWith(tipe)
	proto.Metadata.SetTags(tags)

	return proto
}

type Proto struct {
	Metadata object_metadata.MetadataStruct
}

var _ interfaces.CommandComponentWriter = (*Proto)(nil)

func (proto *Proto) SetFlagDefinitions(f interfaces.CLIFlagDefinitions) {
	proto.Metadata.SetFlagDefinitions(f)
}

func (proto Proto) Equals(metadata object_metadata.MetadataMutable) (ok bool) {
	var okType, okMetadata bool

	if !ids.IsEmpty(proto.Metadata.GetType()) &&
		proto.Metadata.GetType().Equals(metadata.GetType()) {
		okType = true
	}

	if object_metadata.Equaler.Equals(&proto.Metadata, metadata) {
		okMetadata = true
	}

	ok = okType && okMetadata

	return ok
}

func (proto Proto) Make() (object *Transacted) {
	comments.Change("add type")
	comments.Change("add description")
	object = GetTransactedPool().Get()

	proto.Apply(object, genres.Zettel)

	return object
}

func (proto Proto) ApplyType(
	metadataLike object_metadata.GetterMutable,
	genreGetter interfaces.GenreGetter,
) (ok bool) {
	metadata := metadataLike.GetMetadataMutable()

	g := genreGetter.GetGenre()
	ui.Log().Print(metadataLike, g)

	switch g {
	case genres.Zettel, genres.None:
		if ids.IsEmpty(metadata.GetType()) &&
			!ids.IsEmpty(proto.Metadata.GetType()) &&
			!metadata.GetType().Equals(proto.Metadata.GetType()) {
			ok = true
			metadata.GetTypeMutable().ResetWith(proto.Metadata.GetType())
		}
	}

	return ok
}

func (proto Proto) Apply(
	metadataLike object_metadata.GetterMutable,
	genreGetter interfaces.GenreGetter,
) (changed bool) {
	metadata := metadataLike.GetMetadataMutable()

	if proto.ApplyType(metadataLike, genreGetter) {
		changed = true
	}

	if proto.Metadata.GetDescription().WasSet() &&
		!metadata.GetDescription().Equals(proto.Metadata.GetDescription()) {
		changed = true
		metadata.GetDescriptionMutable().ResetWith(proto.Metadata.GetDescription())
	}

	if proto.Metadata.GetTags().Len() > 0 {
		changed = true
	}

	for e := range proto.Metadata.AllTags() {
		errors.PanicIfError(metadata.AddTag(e))
	}

	return changed
}

func (proto Proto) ApplyWithBlobFD(
	metadataGetter object_metadata.GetterMutable,
	blobFD *fd.FD,
) (err error) {
	metadataMutable := metadataGetter.GetMetadataMutable()

	if ids.IsEmpty(metadataMutable.GetType()) &&
		!ids.IsEmpty(proto.Metadata.GetType()) &&
		!metadataMutable.GetType().Equals(proto.Metadata.GetType()) {
		metadataMutable.GetTypeMutable().ResetWith(proto.Metadata.GetType())
	} else {
		// TODO-P4 use konfig
		ext := blobFD.Ext()

		if ext != "" {
			if err = metadataMutable.GetTypeMutable().Set(blobFD.Ext()); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}
	}

	desc := blobFD.FileNameSansExt()

	if proto.Metadata.GetDescription().WasSet() &&
		!metadataMutable.GetDescription().Equals(proto.Metadata.GetDescription()) {
		desc = proto.Metadata.GetDescription().String()
	}

	if err = metadataMutable.GetDescriptionMutable().Set(desc); err != nil {
		err = errors.Wrap(err)
		return err
	}

	for e := range proto.Metadata.AllTags() {
		errors.PanicIfError(metadataMutable.AddTag(e))
	}

	return err
}
