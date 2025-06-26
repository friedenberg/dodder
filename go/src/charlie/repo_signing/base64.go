package repo_signing

import (
	"encoding/base64"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

func SignBase64(key PrivateKey, message []byte) (signature string, err error) {
	var sig []byte

	if sig, err = Sign(key, message); err != nil {
		err = errors.Wrap(err)
		return
	}

	signature = base64.URLEncoding.EncodeToString(sig)

	return
}

func VerifyBase64Signature(
	publicKey PublicKey,
	message []byte,
	signatureBase64 string,
) (err error) {
	var sig []byte

	if sig, err = base64.URLEncoding.DecodeString(signatureBase64); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = VerifySignature(
		publicKey,
		message,
		sig,
	); err != nil {
		err = errors.Wrapf(err, "invalid signature: %q", signatureBase64)
		return
	}

	return
}
