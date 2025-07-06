package query

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
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

func (expObjectId ObjectId) reduce(b *buildState) (err error) {
	if err = expObjectId.GetObjectId().Expand(b.builder.expanders); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO support exact
func (expObjectId ObjectId) ContainsSku(
	objectGetter sku.TransactedGetter,
) (ok bool) {
	object := objectGetter.GetSku()

	metadata := object.GetMetadata()

	method := ids.Contains

	if expObjectId.Exact {
		method = ids.ContainsExactly
	}

	switch expObjectId.GetGenre() {
	case genres.Tag:
		if expObjectId.Exact {
			_, ok = metadata.Cache.TagPaths.All.ContainsObjectIdTagExact(
				expObjectId.GetObjectId(),
			)
		} else {
			_, ok = metadata.Cache.TagPaths.All.ContainsObjectIdTag(
				expObjectId.GetObjectId(),
			)
		}

		if ok {
			return
		}

		return

	case genres.Type:
		if method(metadata.GetType(), expObjectId.GetObjectId()) {
			ok = true
			return
		}

		if e, isExternal := objectGetter.(*sku.Transacted); isExternal {
			if method(e.ExternalType, expObjectId.GetObjectId()) {
				ok = true
				return
			}
		}
	}

	idl := &object.ObjectId

	if !method(idl, expObjectId.GetObjectId()) {
		return
	}

	ok = true

	return
}

func (expObjectId ObjectId) String() string {
	var sb strings.Builder

	if expObjectId.Exact {
		sb.WriteRune('=')
	}

	if expObjectId.Virtual {
		sb.WriteRune('%')
	}

	sb.WriteString(ids.FormattedString(expObjectId.GetObjectId()))

	return sb.String()
}
