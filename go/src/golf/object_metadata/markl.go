package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func (metadata *Metadata) GetBlobDigest() interfaces.MarklId {
	return &metadata.DigBlob
}

func (metadata *Metadata) GetBlobDigestMutable() interfaces.MutableMarklId {
	return &metadata.DigBlob
}

func (metadata *Metadata) GetObjectDigest() interfaces.MarklId {
	return &metadata.digSelf
}

func (metadata *Metadata) GetObjectDigestMutable() interfaces.MutableMarklId {
	return &metadata.digSelf
}

func (metadata *Metadata) GetMotherObjectSig() interfaces.MarklId {
	return &metadata.sigMother
}

func (metadata *Metadata) GetMotherObjectSigMutable() interfaces.MutableMarklId {
	return &metadata.sigMother
}

func (metadata *Metadata) GetRepoPubKey() interfaces.MarklId {
	return metadata.pubRepo
}

func (metadata *Metadata) GetRepoPubKeyMutable() interfaces.MutableMarklId {
	return &metadata.pubRepo
}

func (metadata *Metadata) GetObjectSig() interfaces.MarklId {
	return &metadata.sigRepo
}

func (metadata *Metadata) GetObjectSigMutable() interfaces.MutableMarklId {
	return &metadata.sigRepo
}
