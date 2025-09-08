package markl

import (
	"crypto/ed25519"
	"crypto/rand"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func GeneratePrivateKey(
	rand io.Reader,
	format string,
	tipe string,
	dst interfaces.MutableMarklId,
) (err error) {
	switch tipe {
	default:
		err = errors.Errorf("unsupported type: %q", tipe)
		return

	case TypeIdEd25519Sec:
		var src ed25519.PrivateKey

		if _, src, err = ed25519.GenerateKey(rand); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = dst.SetFormat(format); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = dst.SetMerkleId(tipe, src); err != nil {
			err = errors.Wrap(err)
			return
		}

	case TypeIdAgeSec:
		err = errors.Err501NotImplemented
		return
	}

	return
}

func GetPublicKey(private interfaces.MarklId) (public Id, err error) {
	marklTypeId := private.GetMarklType().GetMarklTypeId()
	switch marklTypeId {
	default:
		err = errors.Errorf(
			"unsupported id: %q. Type: %q",
			private.StringWithFormat(),
			marklTypeId,
		)
		return

	case FormatIdRepoPrivateKeyV1:
		// legacy
		fallthrough

	case TypeIdEd25519Sec:
		if err = public.SetFormat(FormatIdRepoPubKeyV1); err != nil {
			err = errors.Wrap(err)
			return
		}

		pubKeyBytes := ed25519.PrivateKey(private.GetBytes()).Public().(ed25519.PublicKey)

		if err = public.SetMerkleId(TypeIdEd25519Pub, pubKeyBytes); err != nil {
			err = errors.Wrap(err)
			return
		}

	case TypeIdAgeSec:
		if err = public.SetFormat(FormatIdRepoPubKeyV1); err != nil {
			err = errors.Wrap(err)
			return
		}

		pubKeyBytes := ed25519.PrivateKey(private.GetBytes()).Public().(ed25519.PublicKey)

		if err = public.SetMerkleId(TypeIdEd25519Pub, pubKeyBytes); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func MakeNonce(bites []byte, format string) (nonce Id, err error) {
	if format == "" {
		format = FormatIdRequestAuthChallengeV1
	}

	if len(bites) == 0 {
		bites = make([]byte, 32)

		if _, err = rand.Read(bites); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if err = nonce.SetFormat(format); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = nonce.SetMerkleId(
		TypeIdNonce,
		bites,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
