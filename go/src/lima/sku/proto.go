package sku

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/comments"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/golf/fd"
	"code.linenisgreat.com/dodder/go/src/hotel/repo_configs"
	"code.linenisgreat.com/dodder/go/src/juliett/object_metadata"
)

func MakeProto(defaults repo_configs.Defaults) (proto Proto) {
	var tipe ids.Type
	var tags ids.TagSet

	if defaults != nil {
		tipe = defaults.GetDefaultType()
		tags = ids.MakeTagSet(defaults.GetDefaultTags()...)
	}

	proto.Metadata.Type = tipe
	proto.Metadata.SetTags(tags)

	return proto
}

type Proto struct {
	object_metadata.Metadata
}

var _ interfaces.CommandComponentWriter = (*Proto)(nil)

func (proto *Proto) SetFlagDefinitions(f interfaces.CLIFlagDefinitions) {
	proto.Metadata.SetFlagDefinitions(f)
}

func (proto Proto) Equals(metadata object_metadata.IMetadataMutable) (ok bool) {
	var okType, okMetadata bool

	if !ids.IsEmpty(proto.Metadata.Type) &&
		proto.Metadata.Type.Equals(metadata.GetType()) {
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
			!ids.IsEmpty(proto.Metadata.Type) &&
			!metadata.GetType().Equals(proto.Metadata.Type) {
			ok = true
			metadata.GetTypePtr().ResetWith(proto.Metadata.Type)
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

	if proto.Metadata.Description.WasSet() &&
		!metadata.GetDescription().Equals(proto.Metadata.Description) {
		changed = true
		metadata.GetDescriptionMutable().ResetWith(proto.Metadata.GetDescription())
	}

	if proto.Metadata.GetTags().Len() > 0 {
		changed = true
	}

	for e := range proto.Metadata.GetTags().AllPtr() {
		errors.PanicIfError(metadata.AddTagPtr(e))
	}

	return changed
}

func (proto Proto) ApplyWithBlobFD(
	metadataGetter object_metadata.GetterMutable,
	blobFD *fd.FD,
) (err error) {
	metadataMutable := metadataGetter.GetMetadataMutable()

	if ids.IsEmpty(metadataMutable.GetType()) &&
		!ids.IsEmpty(proto.Metadata.Type) &&
		!metadataMutable.GetType().Equals(proto.Metadata.Type) {
		metadataMutable.GetTypePtr().ResetWith(proto.Metadata.Type)
	} else {
		// TODO-P4 use konfig
		ext := blobFD.Ext()

		if ext != "" {
			if err = metadataMutable.GetTypePtr().Set(blobFD.Ext()); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}
	}

	desc := blobFD.FileNameSansExt()

	if proto.Metadata.Description.WasSet() &&
		!metadataMutable.GetDescription().Equals(proto.Metadata.Description) {
		desc = proto.Metadata.Description.String()
	}

	if err = metadataMutable.GetDescriptionMutable().Set(desc); err != nil {
		err = errors.Wrap(err)
		return err
	}

	for e := range proto.Metadata.GetTags().AllPtr() {
		errors.PanicIfError(metadataMutable.AddTagPtr(e))
	}

	return err
}
