package objects

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
)

func (metadata *metadata) GetBlobDigest() interfaces.MarklId {
	return &metadata.DigBlob
}

func (metadata *metadata) GetBlobDigestMutable() interfaces.MarklIdMutable {
	return &metadata.DigBlob
}

func (metadata *metadata) GetObjectDigest() interfaces.MarklId {
	return &metadata.digSelf
}

func (metadata *metadata) GetObjectDigestMutable() interfaces.MarklIdMutable {
	return &metadata.digSelf
}

func (metadata *metadata) GetMotherObjectSig() interfaces.MarklId {
	return &metadata.sigMother
}

func (metadata *metadata) GetMotherObjectSigMutable() interfaces.MarklIdMutable {
	return &metadata.sigMother
}

func (metadata *metadata) GetRepoPubKey() interfaces.MarklId {
	return metadata.pubRepo
}

func (metadata *metadata) GetRepoPubKeyMutable() interfaces.MarklIdMutable {
	return &metadata.pubRepo
}

func (metadata *metadata) GetObjectSig() interfaces.MarklId {
	return &metadata.sigRepo
}

func (metadata *metadata) GetObjectSigMutable() interfaces.MarklIdMutable {
	return &metadata.sigRepo
}
