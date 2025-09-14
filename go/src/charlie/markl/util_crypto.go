package markl

import (
	"crypto/ed25519"
	"crypto/rand"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/zit/go/zit/src/bravo/bech32"
)

func GeneratePrivateKey(
	rand io.Reader,
	purpose string,
	formatId string,
	dst interfaces.MutableMarklId,
) (err error) {
	switch formatId {
	default:
		err = errors.Errorf("unsupported format: %q", formatId)
		return

	case FormatIdEd25519Sec:
		var src ed25519.PrivateKey

		if _, src, err = ed25519.GenerateKey(rand); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = dst.SetPurpose(purpose); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = dst.SetMarklId(formatId, src); err != nil {
			err = errors.Wrap(err)
			return
		}

	case FormatIdAgeX25519Sec:
		var ageId age.Identity

		if err = ageId.GenerateIfNecessary(); err != nil {
			err = errors.Wrap(err)
			return
		}

		bech32String := ageId.String()

		var data []byte

		if _, data, err = bech32.Decode(bech32String); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = dst.SetPurpose(purpose); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = dst.SetMarklId(
			FormatIdAgeX25519Sec,
			data,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func GetPublicKey(private interfaces.MarklId) (public Id, err error) {
	marklTypeId := private.GetMarklFormat().GetMarklFormatId()

	switch marklTypeId {
	default:
		err = errors.Errorf(
			"unsupported id: %q. Type: %q",
			private.StringWithFormat(),
			marklTypeId,
		)
		return

	case PurposeRepoPrivateKeyV1:
		// legacy
		fallthrough

	case FormatIdEd25519Sec:
		if err = public.SetPurpose(PurposeRepoPubKeyV1); err != nil {
			err = errors.Wrap(err)
			return
		}

		privateBytes := private.GetBytes()
		var privateKey ed25519.PrivateKey

		switch len(privateBytes) {
		case ed25519.SeedSize:
			// TODO emit error
			err = errors.Errorf("private key is just seed, not full go ed25519 private key")
			return
			privateKey = ed25519.NewKeyFromSeed(privateBytes)

		case ed25519.PrivateKeySize:
			privateKey = ed25519.PrivateKey(privateBytes)

		default:
			err = errors.Errorf("unsupported key size: %d", len(privateBytes))
			return
		}

		pubKey := privateKey.Public()
		pubKeyBytes := pubKey.(ed25519.PublicKey)

		if err = public.SetMarklId(FormatIdEd25519Pub, pubKeyBytes); err != nil {
			err = errors.Wrap(err)
			return
		}

	case FormatIdAgeX25519Sec:
		if err = public.SetPurpose(PurposeRepoPubKeyV1); err != nil {
			err = errors.Wrap(err)
			return
		}

		// the ed25519 package includes a public key suffix, so we need to
		// reconstruct their version of a private key for a public key value
		privateKey := ed25519.PrivateKey(private.GetBytes())
		pubKeyBytes := privateKey.Public().(ed25519.PublicKey)

		if err = public.SetMarklId(FormatIdEd25519Pub, pubKeyBytes); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func MakeNonce(bites []byte, format string) (nonce Id, err error) {
	if format == "" {
		format = PurposeRequestAuthChallengeV1
	}

	if len(bites) == 0 {
		bites = make([]byte, 32)

		if _, err = rand.Read(bites); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if err = nonce.SetPurpose(format); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = nonce.SetMarklId(
		FormatIdNonce,
		bites,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func GetIOWrapper(
	private interfaces.MarklId,
) (ioWrapper interfaces.IOWrapper, err error) {
	marklType := private.GetMarklFormat()

	if marklType == nil {
		ioWrapper = files.NopeIOWrapper{}
		return
	}

	marklTypeId := marklType.GetMarklFormatId()

	switch marklTypeId {
	default:
		err = errors.Errorf(
			"unsupported id: %q. Type: %q",
			private.StringWithFormat(),
			marklTypeId,
		)

		return

	case FormatIdAgeX25519Sec:
		var ageId age.Identity

		var bech32String []byte

		if bech32String, err = bech32.Encode(
			"AGE-SECRET-KEY-",
			private.GetBytes(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = ageId.Set(string(bech32String)); err != nil {
			err = errors.Wrap(err)
			return
		}

		ioWrapper = &ageId
	}

	return
}
