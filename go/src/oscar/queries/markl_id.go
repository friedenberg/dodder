package queries

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/hotel/objects"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

type MarklId struct {
	Id markl.Id
}

var _ HoistedId = MarklId{}

// TODO add support for abbreviated markl ids
func (marklId MarklId) reduce(buildState *buildState) (err error) {
	return err
}

func (marklId *MarklId) ReadFromSeq(seq doddish.Seq) (err error) {
	coder := markl.MakeIdCoderDoddish(&marklId.Id)

	if err = coder.UnmarshalDoddish(seq); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO support exact
func (marklId MarklId) ContainsSku(
	objectGetter sku.TransactedGetter,
) (ok bool) {
	metadata := objectGetter.GetSku().GetMetadata()

	id := objects.GetMarklIdForPurpose(
		metadata,
		marklId.Id.GetPurpose(),
	)

	return markl.Equals(marklId.Id, id)
}

func (marklId MarklId) String() string {
	return marklId.Id.String()
}

func (marklId MarklId) GetGenre() interfaces.Genre {
	return genres.Blob
}

func (marklId MarklId) IsEmpty() bool {
	return marklId.Id.IsEmpty()
}
