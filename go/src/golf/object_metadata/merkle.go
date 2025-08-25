package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
)

func (metadata *Metadata) GetDigest() interfaces.BlobId {
	return &metadata.digestSelf
}

func (metadata *Metadata) GetDigestMutable() interfaces.MutableGenericBlobId {
	return &metadata.digestSelf
}

func (metadata *Metadata) GetMotherDigest() interfaces.BlobId {
	return &metadata.digestMother
}

func (metadata *Metadata) GetMotherDigestMutable() interfaces.MutableGenericBlobId {
	return &metadata.digestMother
}

func (metadata *Metadata) GetContentSig() interfaces.MerkleId {
	return &metadata.RepoSig
}

func (metadata *Metadata) GetContentSigMutable() interfaces.MutableMerkleId {
	return &metadata.RepoSig
}

func (metadata *Metadata) GetRepoPubkeyValue() blech32.Value {
	return blech32.Value{
		// TODO determine based on object root type
		HRP:  merkle.HRPRepoPubKeyV1,
		Data: metadata.RepoPubkey,
	}
}

func (metadata *Metadata) GetRepoSigValue() blech32.Value {
	return blech32.Value{
		// TODO determine based on object root type
		HRP:  merkle.HRPRepoSigV1,
		Data: metadata.RepoSig.GetBytes(),
	}
}
