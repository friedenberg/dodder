package sku

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/external_state"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

type Transacted struct {
	ObjectId ids.ObjectId
	Metadata object_metadata.Metadata

	ExternalType ids.Type

	// TODO add support for querying the below
	RepoId           ids.RepoId
	State            external_state.State
	ExternalObjectId ids.ExternalObjectId
}

var _ object_metadata.GetterMutable = &Transacted{}

func (transacted *Transacted) GetSkuExternal() *Transacted {
	return transacted
}

func (transacted *Transacted) GetRepoId() ids.RepoId {
	return transacted.RepoId
}

func (transacted *Transacted) GetExternalObjectId() ids.ExternalObjectIdLike {
	return &transacted.ExternalObjectId
}

func (transacted *Transacted) GetExternalState() external_state.State {
	return transacted.State
}

func (transacted *Transacted) CloneTransacted() (cloned *Transacted) {
	cloned = GetTransactedPool().Get()
	TransactedResetter.ResetWith(cloned, transacted)
	return cloned
}

func (transacted *Transacted) GetSku() *Transacted {
	return transacted
}

func (transacted *Transacted) SetFromTransacted(other *Transacted) (err error) {
	TransactedResetter.ResetWith(transacted, other)

	return err
}

func (transacted *Transacted) Less(other *Transacted) bool {
	less := transacted.GetTai().Less(other.GetTai())

	return less
}

func (transacted *Transacted) GetTags() ids.TagSet {
	return transacted.Metadata.GetTags()
}

func (transacted *Transacted) AddTagPtr(tag *ids.Tag) (err error) {
	if transacted.ObjectId.GetGenre() == genres.Tag &&
		strings.HasPrefix(transacted.ObjectId.String(), tag.String()) {
		return err
	}

	tagKey := transacted.Metadata.Cache.GetImplicitTags().KeyPtr(tag)

	if transacted.Metadata.Cache.GetImplicitTags().ContainsKey(tagKey) {
		return err
	}

	if err = transacted.GetMetadataMutable().AddTagPtr(tag); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (transacted *Transacted) AddTagPtrFast(tag *ids.Tag) (err error) {
	if err = transacted.GetMetadataMutable().AddTagPtrFast(tag); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (transacted *Transacted) GetType() ids.Type {
	return transacted.Metadata.Type
}

func (transacted *Transacted) GetMetadataMutable() *object_metadata.Metadata {
	return &transacted.Metadata
}

func (transacted *Transacted) GetTai() ids.Tai {
	return transacted.Metadata.GetTai()
}

func (transacted *Transacted) SetTai(tai ids.Tai) {
	transacted.GetMetadataMutable().Tai = tai
}

func (transacted *Transacted) GetObjectId() *ids.ObjectId {
	return &transacted.ObjectId
}

func (transacted *Transacted) SetObjectIdLike(
	objectIdLike interfaces.ObjectId,
) (err error) {
	if err = transacted.ObjectId.SetWithIdLike(objectIdLike); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (transacted *Transacted) EqualsAny(other any) (ok bool) {
	return values.Equals(transacted, other)
}

func (transacted *Transacted) Equals(other *Transacted) (ok bool) {
	if transacted.GetObjectId().String() != other.GetObjectId().String() {
		return ok
	}

	// TODO-P2 determine why object shas in import test differed
	// if !a.Metadata.Sha().Equals(b.Metadata.Sha()) {
	// 	return
	// }

	if !transacted.Metadata.Equals(&other.Metadata) {
		return ok
	}

	return true
}

func (transacted *Transacted) GetGenre() interfaces.Genre {
	return transacted.ObjectId.GetGenre()
}

func (transacted *Transacted) IsNew() bool {
	return transacted.Metadata.GetMotherObjectSig().IsNull()
}

func (transacted *Transacted) SetDormant(v bool) {
	transacted.Metadata.Cache.Dormant.SetBool(v)
}

func (transacted *Transacted) GetObjectDigest() interfaces.MarklId {
	return transacted.GetMetadataMutable().GetObjectDigest()
}

func (transacted *Transacted) GetBlobDigest() interfaces.MarklId {
	return transacted.Metadata.GetBlobDigest()
}

func (transacted *Transacted) SetBlobDigest(
	merkleId interfaces.MarklId,
) (err error) {
	if err = transacted.Metadata.GetBlobDigestMutable().SetMarklId(
		merkleId.GetMarklFormat().GetMarklFormatId(),
		merkleId.GetBytes(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (transacted *Transacted) GetKey() string {
	return ids.FormattedString(transacted.GetObjectId())
}

func (transacted *Transacted) GetStringProbeKeys() map[string]string {
	return map[string]string{
		"objectId":     transacted.GetObjectId().String(),
		"objectId+tai": transacted.GetObjectId().String() + transacted.GetTai().String(),
	}
}

func (transacted *Transacted) AllProbeIds(
	hashType markl.FormatHash,
) interfaces.Seq[ids.ProbeId] {
	return func(yield func(ids.ProbeId) bool) {
		for key, value := range transacted.GetStringProbeKeys() {
			id, repool := hashType.GetMarklIdForString(value)

			probeId := ids.ProbeId{
				Key: key,
				Id:  id,
			}

			if !yield(probeId) {
				repool()
				return
			}

			repool()
		}

		{
			probeId := ids.ProbeId{
				Key: markl.PurposeObjectDigestV1,
				Id:  transacted.Metadata.GetObjectDigest(),
			}

			if !yield(probeId) {
				return
			}
		}

		{
			probeId := ids.ProbeId{
				Key: markl.PurposeObjectDigestV2,
				Id:  transacted.Metadata.GetObjectDigest(),
			}

			if !yield(probeId) {
				return
			}
		}

		{
			probeId := ids.ProbeId{
				Key: markl.PurposeV5MetadataDigestWithoutTai,
				Id:  transacted.Metadata.SelfWithoutTai,
			}

			if !yield(probeId) {
				return
			}
		}

		{
			probeId := ids.ProbeId{
				Key: markl.PurposeObjectSigV1,
				Id:  transacted.Metadata.GetObjectSig(),
			}

			if !yield(probeId) {
				return
			}
		}
	}
}
