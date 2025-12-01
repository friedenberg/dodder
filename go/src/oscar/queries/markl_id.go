package queries

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

type MarklId struct {
	Id markl.Id
}

// TODO add support for abbreviated markl ids
func (marklId MarklId) reduce(b *buildState) (err error) {
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

	var id interfaces.MarklId

	purposeType := markl.GetPurpose(marklId.Id.GetPurpose()).GetPurposeType()

	switch purposeType {

	case markl.PurposeTypeBlobDigest:
		id = metadata.GetBlobDigest()

	case markl.PurposeTypeObjectMotherSig:
		id = metadata.GetMotherObjectSig()

	case markl.PurposeTypeObjectSig:
		id = metadata.GetObjectSig()

	case markl.PurposeTypeRepoPubKey:
		id = metadata.GetRepoPubKey()

	default:
		panic(fmt.Sprintf("unsupported purpose type: %q", purposeType))
	}

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
