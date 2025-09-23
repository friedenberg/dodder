package query

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/doddish"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type ObjectId struct {
	Exact   bool
	Virtual bool
	Debug   bool

	*ids.ObjectId
}

func (objectId ObjectId) reduce(b *buildState) (err error) {
	if err = objectId.GetObjectId().Expand(b.builder.expanders); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (objectId *ObjectId) ReadFromSeq(seq doddish.Seq) (err error) {
	ok, left, _ := seq.MatchEnd(doddish.TokenMatcherOp(doddish.OpExact))

	if ok {
		objectId.Exact = true
		seq = left
	}

	if err = objectId.GetObjectId().ReadFromSeq(seq); err != nil {
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
	case genres.Tag:
		if objectId.Exact {
			_, ok = metadata.Cache.TagPaths.All.ContainsObjectIdTagExact(
				objectId.GetObjectId(),
			)
		} else {
			_, ok = metadata.Cache.TagPaths.All.ContainsObjectIdTag(
				objectId.GetObjectId(),
			)
		}

		if ok {
			return ok
		}

		return ok

	case genres.Type:
		if method(metadata.GetType(), objectId.GetObjectId()) {
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
