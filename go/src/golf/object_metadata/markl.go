package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
)

func (metadata *metadata) GetBlobDigest() interfaces.MarklId {
	return &metadata.DigBlob
}

func (metadata *metadata) GetBlobDigestMutable() interfaces.MutableMarklId {
	return &metadata.DigBlob
}

func (metadata *metadata) GetObjectDigest() interfaces.MarklId {
	return &metadata.digSelf
}

func (metadata *metadata) GetObjectDigestMutable() interfaces.MutableMarklId {
	return &metadata.digSelf
}

func (metadata *metadata) GetMotherObjectSig() interfaces.MarklId {
	return &metadata.sigMother
}

func (metadata *metadata) GetMotherObjectSigMutable() interfaces.MutableMarklId {
	return &metadata.sigMother
}

func (metadata *metadata) GetRepoPubKey() interfaces.MarklId {
	return metadata.pubRepo
}

func (metadata *metadata) GetRepoPubKeyMutable() interfaces.MutableMarklId {
	return &metadata.pubRepo
}

func (metadata *metadata) GetObjectSig() interfaces.MarklId {
	return &metadata.sigRepo
}

func (metadata *metadata) GetObjectSigMutable() interfaces.MutableMarklId {
	return &metadata.sigRepo
}
