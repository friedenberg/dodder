package queries

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/hotel/objects"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

type ObjectId struct {
	Exact   bool
	Virtual bool
	Debug   bool

	*ids.ObjectId

	marklId markl.Id
}

var _ ObjectId = ObjectId{}

func (objectId *ObjectId) reduce(buildState *buildState) (err error) {
	if err = ids.Expand(
		objectId.GetObjectId(),
		buildState.builder.expanders,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if objectId.GetGenre() == genres.Blob {
		if err = objectId.marklId.Set(objectId.GetObjectId().String()); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (objectId *ObjectId) ReadFromSeq(seq doddish.Seq) (err error) {
	ok, left, _ := seq.MatchEnd(doddish.TokenMatcherOp(doddish.OpExact))

	if ok {
		objectId.Exact = true
		seq = left
	}

	if err = objectId.GetObjectId().SetWithSeq(seq); err != nil {
		if errors.Is(err, doddish.ErrUnsupportedSeq{}) {
			err = errors.BadRequest(err)
		} else {
			err = errors.Wrap(err)
		}

		return err
	}

	return err
}

// TODO support exact
func (objectId ObjectId) ContainsSku(
	objectGetter sku.TransactedGetter,
) (ok bool) {
	object := objectGetter.GetSku()

	metadata := object.GetMetadata()

	method := ids.Contains

	if objectId.Exact {
		method = ids.ContainsExactly
	}

	switch objectId.GetGenre() {

	case genres.Blob:
		purposeId := objectId.marklId.GetPurposeId()

		id := objects.GetMarklIdForPurpose(metadata, purposeId)

		return markl.Equals(objectId.marklId, id)

	case genres.Tag:
		if objectId.Exact {
			_, ok = metadata.GetIndex().GetTagPaths().All.ContainsObjectIdTagExact(
				objectId.GetObjectId(),
			)
		} else {
			_, ok = metadata.GetIndex().GetTagPaths().All.ContainsObjectIdTag(
				objectId.GetObjectId(),
			)
		}

		if ok {
			return ok
		}

		return ok

	case genres.Type:
		if method(metadata.GetType().ToType(), objectId.GetObjectId()) {
			ok = true
			return ok
		}

		if e, isExternal := objectGetter.(*sku.Transacted); isExternal {
			if method(e.ExternalType, objectId.GetObjectId()) {
				ok = true
				return ok
			}
		}
	}

	idl := &object.ObjectId

	if !method(idl, objectId.GetObjectId()) {
		return ok
	}

	ok = true

	return ok
}

func (objectId ObjectId) String() string {
	var sb strings.Builder

	if objectId.Exact {
		sb.WriteRune('=')
	}

	if objectId.Virtual {
		sb.WriteRune('%')
	}

	sb.WriteString(ids.FormattedString(objectId.GetObjectId()))

	return sb.String()
}
