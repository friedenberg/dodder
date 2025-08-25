package remote_http

import (
	"bytes"
	"crypto/rand"
	"net/http"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
)

const (
	headerChallengeNonce    = "X-Dodder-Challenge-Nonce"
	headerChallengeResponse = "X-Dodder-Challenge-Response"
	headerRepoPublicKey     = "X-Dodder-Repo-Public_Key"
	headerRepoSig           = "X-Dodder-Repo-Sig"
)

type RoundTripperBufioWrappedSigner struct {
	merkle.PublicKey
	roundTripperBufio
}

// TODO extract signing into an agnostic middleware
func (roundTripper *RoundTripperBufioWrappedSigner) RoundTrip(
	request *http.Request,
) (response *http.Response, err error) {
	nonceBytes := make([]byte, 32)

	if _, err = rand.Read(nonceBytes); err != nil {
		err = errors.Wrap(err)
		return
	}

	nonce := blech32.Value{
		HRP:  merkle.HRPRequestAuthChallengeV1,
		Data: nonceBytes,
	}

	request.Header.Add(headerChallengeNonce, nonce.String())

	if response, err = roundTripper.roundTripperBufio.RoundTrip(
		request,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	sigString := response.Header.Get(headerChallengeResponse)

	if sigString == "" {
		err = errors.Errorf("signature empty or not provided")
		return
	}

	var sig blech32.Value

	if sig, err = blech32.MakeValueWithExpectedHRP(
		merkle.HRPRequestAuthResponseV1,
		sigString,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	pubkeyString := response.Header.Get(headerRepoPublicKey)

	var pubkey blech32.Value

	if pubkey, err = blech32.MakeValueWithExpectedHRP(
		merkle.HRPRepoPubKeyV1,
		pubkeyString,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if roundTripper.PublicKey.IsEmpty() {
		// TODO present prompt to user for TOFU
	} else {
		if !bytes.Equal(roundTripper.PublicKey.GetBytes(), pubkey.Data) {
			err = errors.Errorf(
				"expected pubkey %q but got %q",
				roundTripper.PublicKey.GetBytes(),
				pubkey.Data,
			)

			return
		}
	}

	if err = merkle.VerifySignature(
		pubkey.Data,
		nonceBytes,
		sig.Data,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
