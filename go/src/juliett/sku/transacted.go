package sku

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/external_state"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
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

func (t *Transacted) GetSkuExternal() *Transacted {
	return t
}

func (t *Transacted) GetRepoId() ids.RepoId {
	return t.RepoId
}

func (t *Transacted) GetExternalObjectId() ids.ExternalObjectIdLike {
	return &t.ExternalObjectId
}

func (t *Transacted) GetExternalState() external_state.State {
	return t.State
}

func (transacted *Transacted) CloneTransacted() (cloned *Transacted) {
	cloned = GetTransactedPool().Get()
	TransactedResetter.ResetWith(cloned, transacted)
	return
}

func (t *Transacted) GetSku() *Transacted {
	return t
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

func (s *Transacted) GetGenre() interfaces.Genre {
	return s.ObjectId.GetGenre()
}

func (s *Transacted) IsNew() bool {
	return s.Metadata.Mutter().IsNull()
}

func (s *Transacted) CalculateObjectShaDebug() (err error) {
	return s.calculateObjectSha(true)
}

// TODO replace this with repo signatures
func (s *Transacted) CalculateObjectShas() (err error) {
	return s.calculateObjectSha(false)
}

func (transacted *Transacted) makeShaCalcFunc(
	f func(object_inventory_format.FormatGeneric, object_inventory_format.FormatterContext) (*sha.Sha, error),
	of object_inventory_format.FormatGeneric,
	sh *sha.Sha,
) errors.FuncErr {
	return func() (err error) {
		var actual *sha.Sha

		if actual, err = f(
			of,
			transacted,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer digests.PutDigest(actual)

		sh.ResetWith(actual)

		return
	}
}

func (transacted *Transacted) calculateObjectSha(debug bool) (err error) {
	f := object_inventory_format.GetShaForContext

	if debug {
		f = object_inventory_format.GetShaForContextDebug
	}

	wg := errors.MakeWaitGroupParallel()

	wg.Do(
		transacted.makeShaCalcFunc(
			f,
			object_inventory_format.Formats.MetadataObjectIdParent(),
			transacted.Metadata.Sha(),
		),
	)

	wg.Do(
		transacted.makeShaCalcFunc(
			f,
			object_inventory_format.Formats.Metadata(),
			&transacted.Metadata.SelfMetadata,
		),
	)

	wg.Do(
		transacted.makeShaCalcFunc(
			f,
			object_inventory_format.Formats.MetadataSansTai(),
			&transacted.Metadata.SelfMetadataWithoutTai,
		),
	)

	return wg.GetError()
}

func (transacted *Transacted) SetDormant(v bool) {
	transacted.Metadata.Cache.Dormant.SetBool(v)
}

func (transacted *Transacted) SetObjectSha(v interfaces.Digest) (err error) {
	return transacted.GetMetadata().Sha().SetDigest(v)
}

func (transacted *Transacted) GetObjectSha() interfaces.Digest {
	return transacted.GetMetadata().Sha()
}

func (transacted *Transacted) GetBlobSha() interfaces.Digest {
	return &transacted.Metadata.Blob
}

func (transacted *Transacted) SetBlobSha(sh interfaces.Digest) error {
	return transacted.Metadata.Blob.SetDigest(sh)
}

func (transacted *Transacted) GetKey() string {
	return ids.FormattedString(transacted.GetObjectId())
}

func (transacted *Transacted) Sign(
	config genesis_configs.ConfigPrivate,
) (err error) {
	transacted.Metadata.RepoPubkey = config.GetPublicKey()

	sh := sha.MustWithDigester(transacted.GetTai())
	defer digests.PutDigest(sh)

	if transacted.Metadata.RepoSig, err = repo_signing.Sign(
		config.GetPrivateKey(),
		sh.GetBytes(),
	); err != nil {
		err = errors.Wrap(err)
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
