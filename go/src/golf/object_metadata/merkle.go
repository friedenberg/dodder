package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func (metadata *Metadata) GetBlobDigest() interfaces.BlobId {
	return &metadata.DigBlob
}

func (metadata *Metadata) GetBlobDigestMutable() interfaces.MutableBlobId {
	return &metadata.DigBlob
}

func (metadata *Metadata) GetObjectDigest() interfaces.BlobId {
	return &metadata.digSelf
}

func (metadata *Metadata) GetObjectDigestMutable() interfaces.MutableBlobId {
	return &metadata.digSelf
}

func (metadata *Metadata) GetMotherObjectSig() interfaces.BlobId {
	return &metadata.sigMother
}

func (metadata *Metadata) GetMotherObjectSigMutable() interfaces.MutableBlobId {
	return &metadata.sigMother
}

func (metadata *Metadata) GetRepoPubKey() interfaces.BlobId {
	return metadata.pubRepo
}

func (metadata *Metadata) GetRepoPubKeyMutable() interfaces.MutableBlobId {
	return &metadata.pubRepo
}

func (metadata *Metadata) GetObjectSig() interfaces.BlobId {
	return &metadata.sigRepo
}

func (metadata *Metadata) GetObjectSigMutable() interfaces.MutableBlobId {
	return &metadata.sigRepo
}
