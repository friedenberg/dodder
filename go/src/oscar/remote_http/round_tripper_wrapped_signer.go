package remote_http

import (
	"bytes"
	"net/http"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
)

const (
	headerChallengeNonce    = "X-Dodder-Challenge-Nonce"
	headerChallengeResponse = "X-Dodder-Challenge-Response"
	headerRepoPublicKey     = "X-Dodder-Repo-Public_Key"
	headerRepoSig           = "X-Dodder-Repo-Sig"
)

type RoundTripperBufioWrappedSigner struct {
	PublicKey markl.Id
	roundTripperBufio
}

// TODO extract signing into an agnostic middleware
func (roundTripper *RoundTripperBufioWrappedSigner) RoundTrip(
	request *http.Request,
) (response *http.Response, err error) {
	var nonce markl.Id

	if nonce, err = markl.MakeNonce(
		nil,
		markl.FormatIdRequestAuthChallengeV1,
	); err != nil {
		err = errors.Wrap(err)
		return
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

	var sig markl.Id

	if err = sig.Set(sigString); err != nil {
		err = errors.Wrap(err)
		return
	}

	pubkeyString := response.Header.Get(headerRepoPublicKey)

	var pubkey markl.Id

	if err = pubkey.Set(
		pubkeyString,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if roundTripper.PublicKey.IsEmpty() {
		// TODO present prompt to user for TOFU
	} else {
		if !bytes.Equal(roundTripper.PublicKey.GetBytes(), pubkey.GetBytes()) {
			err = errors.Errorf(
				"expected pubkey %q but got %q",
				roundTripper.PublicKey.GetBytes(),
				pubkey.GetBytes(),
			)

			return
		}
	}

	if err = markl.Verify(
		pubkey,
		nonce,
		sig,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
