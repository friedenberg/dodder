package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func (metadata *Metadata) GetBlobDigest() interfaces.MerkleId {
	return &metadata.Blob
}

func (metadata *Metadata) GetBlobDigestMutable() interfaces.MutableMerkleId {
	return &metadata.Blob
}

func (metadata *Metadata) GetObjectDigest() interfaces.MerkleId {
	return &metadata.digSelf
}

func (metadata *Metadata) GetObjectDigestMutable() interfaces.MutableMerkleId {
	return &metadata.digSelf
}

func (metadata *Metadata) GetMotherObjectDigest() interfaces.MerkleId {
	return &metadata.digMother
}

func (metadata *Metadata) GetMotherObjectDigestMutable() interfaces.MutableMerkleId {
	return &metadata.digMother
}

func (metadata *Metadata) GetRepoPubKey() interfaces.MerkleId {
	return metadata.pubRepo
}

func (metadata *Metadata) GetPubKeyMutable() interfaces.MutableMerkleId {
	return &metadata.pubRepo
}

func (metadata *Metadata) GetObjectSig() interfaces.MerkleId {
	return &metadata.sigRepo
}

func (metadata *Metadata) GetObjectSigMutable() interfaces.MutableMerkleId {
	return &metadata.sigRepo
}
