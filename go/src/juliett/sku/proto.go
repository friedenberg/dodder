package sku

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/todo"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/repo_config_blobs"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

func MakeProto(defaults repo_config_blobs.Defaults) (proto Proto) {
	var tipe ids.Type
	var tags ids.TagSet

	if defaults != nil {
		tipe = defaults.GetType()
		tags = ids.MakeTagSet(defaults.GetTags()...)
	}

	proto.Metadata.Type = tipe
	proto.Metadata.SetTags(tags)

	return
}

type Proto struct {
	object_metadata.Metadata
}

func (pz *Proto) SetFlagSet(f *flag.FlagSet) {
	pz.Metadata.SetFlagSet(f)
}

func (pz Proto) Equals(z *object_metadata.Metadata) (ok bool) {
	var okTyp, okMet bool

	if !ids.IsEmpty(pz.Metadata.Type) &&
		pz.Metadata.Type.Equals(z.GetType()) {
		okTyp = true
	}

	if pz.Metadata.Equals(z) {
		okMet = true
	}

	ok = okTyp && okMet

	return
}

func (pz Proto) Make() (z *Transacted) {
	todo.Change("add type")
	todo.Change("add description")
	z = GetTransactedPool().Get()

	pz.Apply(z, genres.Zettel)

	return
}

func (proto Proto) ApplyType(
	metadataLike object_metadata.MetadataLike,
	genreGetter interfaces.GenreGetter,
) (ok bool) {
	metadata := metadataLike.GetMetadata()

	g := genreGetter.GetGenre()
	ui.Log().Print(metadataLike, g)

	switch g {
	case genres.Zettel, genres.None:
		if ids.IsEmpty(metadata.GetType()) &&
			!ids.IsEmpty(proto.Metadata.Type) &&
			!metadata.GetType().Equals(proto.Metadata.Type) {
			ok = true
			metadata.Type = proto.Metadata.Type
		}
	}

	return
}

func (proto Proto) Apply(
	metadataLike object_metadata.MetadataLike,
	genreGetter interfaces.GenreGetter,
) (changed bool) {
	metadata := metadataLike.GetMetadata()

	if proto.ApplyType(metadataLike, genreGetter) {
		changed = true
	}

	if proto.Metadata.Description.WasSet() &&
		!metadata.Description.Equals(proto.Metadata.Description) {
		changed = true
		metadata.Description = proto.Metadata.Description
	}

	if proto.Metadata.GetTags().Len() > 0 {
		changed = true
	}

	errors.PanicIfError(proto.Metadata.GetTags().EachPtr(metadata.AddTagPtr))

	return
}

func (pz Proto) ApplyWithBlobFD(
	ml object_metadata.MetadataLike,
	blobFD *fd.FD,
) (err error) {
	z := ml.GetMetadata()

	if ids.IsEmpty(z.GetType()) &&
		!ids.IsEmpty(pz.Metadata.Type) &&
		!z.GetType().Equals(pz.Metadata.Type) {
		z.Type = pz.Metadata.Type
	} else {
		// TODO-P4 use konfig
		ext := blobFD.Ext()

		if ext != "" {
			if err = z.Type.Set(blobFD.Ext()); err != nil {
				err = errors.Wrap(err)
				return
			}
		}
	}

	desc := blobFD.FileNameSansExt()

	if pz.Metadata.Description.WasSet() &&
		!z.Description.Equals(pz.Metadata.Description) {
		desc = pz.Metadata.Description.String()
	}

	if err = z.Description.Set(desc); err != nil {
		err = errors.Wrap(err)
		return
	}

	errors.PanicIfError(pz.Metadata.GetTags().EachPtr(z.AddTagPtr))

	return
}
