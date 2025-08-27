package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
)

func (transacted *Transacted) CalculateObjectDigestsAndMerkleUsingObject() (err error) {
	return transacted.CalculateObjectDigestsAndMerkleUsingRepoPubKey(
		transacted.Metadata.GetRepoPubKey().GetBytes(),
	)
}

func (transacted *Transacted) CalculateObjectDigestsAndMerkleUsingRepoPubKey(
	pubKey merkle.PublicKey,
) (err error) {
	pubKeyMutable := transacted.Metadata.GetRepoPubKeyMutable()

	if pubKeyMutable.IsNull() {
		if err = pubKeyMutable.SetMerkleId(
			merkle.HRPRepoPubKeyV1,
			pubKey,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		if err = merkle_ids.MakeErrNotEqualBytes(
			pubKey,
			pubKeyMutable.GetBytes(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return transacted.calculateObjectSha(false, true)
}
