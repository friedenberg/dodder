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

	if err = transacted.GetMetadata().AddTagPtr(tag); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (transacted *Transacted) AddTagPtrFast(tag *ids.Tag) (err error) {
	if err = transacted.GetMetadata().AddTagPtrFast(tag); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (transacted *Transacted) GetType() ids.Type {
	return transacted.Metadata.Type
}

func (transacted *Transacted) GetMetadata() *object_metadata.Metadata {
	return &transacted.Metadata
}

func (transacted *Transacted) GetTai() ids.Tai {
	return transacted.Metadata.GetTai()
}

func (transacted *Transacted) SetTai(tai ids.Tai) {
	transacted.GetMetadata().Tai = tai
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
	return transacted.GetMetadata().GetObjectDigest()
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

func (transacted *Transacted) GetProbeKeys() map[string]string {
	return map[string]string{
		"objectId":     transacted.GetObjectId().String(),
		"objectId+tai": transacted.GetObjectId().String() + transacted.GetTai().String(),
	}
}

type ProbeId struct {
	Key string
	Id  interfaces.MarklId
}

func (transacted *Transacted) AllProbeIds() interfaces.Seq[ProbeId] {
	return transacted.allProbeIds(markl.FormatHashSha256)
}

func (transacted *Transacted) allProbeIds(
	hashType markl.FormatHash,
) interfaces.Seq[ProbeId] {
	return func(yield func(ProbeId) bool) {
		for key, value := range transacted.GetProbeKeys() {
			id, repool := hashType.GetMarklIdForString(value)

			probeId := ProbeId{
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
			id, repool := hashType.GetMarklIdForMarklId(
				transacted.Metadata.GetObjectSig(),
			)

			probeId := ProbeId{
				Key: markl.PurposeObjectDigestV1,
				Id:  id,
			}

			if !yield(probeId) {
				repool()
				return
			}
			repool()
		}

		{
			id, repool := hashType.GetMarklIdForMarklId(
				transacted.Metadata.GetObjectSig(),
			)

			probeId := ProbeId{
				Key: markl.PurposeV5MetadataDigestWithoutTai,
				Id:  id,
			}

			if !yield(probeId) {
				repool()
				return
			}
			repool()
		}

		{
			id, repool := hashType.GetMarklIdForMarklId(
				transacted.Metadata.GetObjectSig(),
			)

			probeId := ProbeId{
				Key: "object-sig",
				Id:  id,
			}

			if !yield(probeId) {
				repool()
				return
			}

			repool()
		}
	}
}
