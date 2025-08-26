package sku

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/external_state"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
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
	return
}

func (transacted *Transacted) GetSku() *Transacted {
	return transacted
}

func (transacted *Transacted) SetFromTransacted(other *Transacted) (err error) {
	TransactedResetter.ResetWith(transacted, other)

	return
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
		return
	}

	tagKey := transacted.Metadata.Cache.GetImplicitTags().KeyPtr(tag)

	if transacted.Metadata.Cache.GetImplicitTags().ContainsKey(tagKey) {
		return
	}

	if err = transacted.GetMetadata().AddTagPtr(tag); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) AddTagPtrFast(tag *ids.Tag) (err error) {
	if err = transacted.GetMetadata().AddTagPtrFast(tag); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
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
		return
	}

	return
}

func (transacted *Transacted) EqualsAny(other any) (ok bool) {
	return values.Equals(transacted, other)
}

func (transacted *Transacted) Equals(other *Transacted) (ok bool) {
	if transacted.GetObjectId().String() != other.GetObjectId().String() {
		return
	}

	// TODO-P2 determine why object shas in import test differed
	// if !a.Metadata.Sha().Equals(b.Metadata.Sha()) {
	// 	return
	// }

	if !transacted.Metadata.Equals(&other.Metadata) {
		return
	}

	return true
}

func (transacted *Transacted) GetGenre() interfaces.Genre {
	return transacted.ObjectId.GetGenre()
}

func (transacted *Transacted) IsNew() bool {
	return transacted.Metadata.GetMotherObjectDigest().IsNull()
}

// TODO replace this with repo signatures
// TODO include repo pubkey
func (transacted *Transacted) CalculateObjectDigests() (err error) {
	return transacted.calculateObjectSha(false)
}

func (transacted *Transacted) makeDigestCalcFunc(
	funkMakeBlobId func(object_inventory_format.Format, object_inventory_format.FormatterContext) (interfaces.BlobId, error),
	objectFormat object_inventory_format.Format,
	digest interfaces.MutableMerkleId,
) errors.FuncErr {
	return func() (err error) {
		var actual interfaces.BlobId

		if actual, err = funkMakeBlobId(
			objectFormat,
			transacted,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer merkle_ids.PutBlobId(actual)

		if err = digest.SetMerkleId(merkle.HRPObjectDigestSha256V1, actual.GetBytes()); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}
}

func (transacted *Transacted) makeShaCalcFunc(
	f func(object_inventory_format.Format, object_inventory_format.FormatterContext) (interfaces.BlobId, error),
	objectFormat object_inventory_format.Format,
	sh *sha.Sha,
) errors.FuncErr {
	return func() (err error) {
		var actual interfaces.BlobId

		if actual, err = f(
			objectFormat,
			transacted,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer merkle_ids.PutBlobId(actual)

		if err = sh.SetDigest(actual); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}
}

func (transacted *Transacted) calculateObjectSha(debug bool) (err error) {
	f := object_inventory_format.GetDigestForContext

	if debug {
		f = object_inventory_format.GetDigestForContextDebug
	}

	wg := errors.MakeWaitGroupParallel()
	wg.Do(transacted.calculateObjectDigest)

	wg.Do(
		transacted.makeShaCalcFunc(
			f,
			object_inventory_format.FormatsV5MetadataSansTai,
			&transacted.Metadata.SelfWithoutTai,
		),
	)

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) SetDormant(v bool) {
	transacted.Metadata.Cache.Dormant.SetBool(v)
}

func (transacted *Transacted) GetObjectDigest() interfaces.MerkleId {
	return transacted.GetMetadata().GetObjectDigest()
}

// func (transacted *Transacted) GetBlobDigest() interfaces.MerkleId {
// 	return transacted.Metadata.GetBlobDigest()
// }

func (transacted *Transacted) GetBlobDigest() interfaces.BlobId {
	return &transacted.Metadata.Blob
}

func (transacted *Transacted) SetBlobDigest(
	merkleId interfaces.MerkleId,
) (err error) {
	if err = transacted.Metadata.GetBlobDigestMutable().SetMerkleId(
		merkleId.GetType(),
		merkleId.GetBytes(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) GetKey() string {
	return ids.FormattedString(transacted.GetObjectId())
}

func (transacted *Transacted) calculateObjectDigest() (err error) {
	if err = merkle_ids.MakeErrIsNull(
		transacted.Metadata.GetRepoPubKey(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = transacted.makeDigestCalcFunc(
		object_inventory_format.GetDigestForContext,
		object_inventory_format.FormatV11ObjectDigest,
		transacted.Metadata.GetObjectDigestMutable(),
	)(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO turn into proper merkle tree
func (transacted *Transacted) Sign(
	config genesis_configs.ConfigPrivate,
) (err error) {
	transacted.CalculateObjectDigests()

	if err = merkle_ids.MakeErrIsNull(
		transacted.Metadata.GetRepoPubKey()); err != nil {
		err = errors.Wrap(err)
		return
	}

	privateKey := config.GetPrivateKey()

	var bites []byte

	if bites, err = merkle.Sign(
		privateKey,
		transacted.Metadata.GetObjectDigest().GetBytes(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = transacted.Metadata.GetObjectSigMutable().SetMerkleId(
		merkle.HRPRepoSigV1,
		bites,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) Verify() (err error) {
	if err = merkle_ids.MakeErrIsNull(
		transacted.Metadata.GetRepoPubKey(),
	); err != nil {
		err = errors.Wrapf(
			err,
			"Object: %s, Fields: %#v",
			String(transacted),
			transacted.Metadata.Fields,
		)
		return
	}

	if err = merkle_ids.MakeErrIsNull(
		transacted.Metadata.GetObjectSig(),
	); err != nil {
		err = errors.Wrapf(
			err,
			"Object: %s, Fields: %#v",
			String(transacted),
			transacted.Metadata.Fields,
		)
		return
	}

	if err = transacted.CalculateObjectDigests(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = merkle.VerifySignature(
		transacted.Metadata.GetRepoPubKey().GetBytes(),
		transacted.Metadata.GetObjectDigestMutable().GetBytes(),
		transacted.Metadata.GetObjectSig().GetBytes(),
	); err != nil {
		err = errors.Wrapf(
			err,
			"Sku: %s, Tags %s",
			String(transacted),
			quiter.StringCommaSeparated(transacted.Metadata.GetTags()),
		)
		return
	}

	return
}

type transactedLessorTaiOnly struct{}

func (transactedLessorTaiOnly) Less(a, b *Transacted) bool {
	return a.GetTai().Less(b.GetTai())
}

func (transactedLessorTaiOnly) LessPtr(a, b *Transacted) bool {
	return a.GetTai().Less(b.GetTai())
}

type transactedLessorStable struct{}

func (transactedLessorStable) Less(a, b *Transacted) bool {
	if result := a.GetTai().SortCompare(b.GetTai()); !result.Equal() {
		return result.Less()
	}

	return a.GetObjectId().String() < b.GetObjectId().String()
}

func (transactedLessorStable) LessPtr(a, b *Transacted) bool {
	return a.GetTai().Less(b.GetTai())
}

type transactedEqualer struct{}

func (transactedEqualer) Equals(a, b *Transacted) bool {
	return a.Equals(b)
}
