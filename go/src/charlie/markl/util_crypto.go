package markl

import (
	"crypto/ed25519"
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
	var formatSec FormatSec

	if formatSec, err = GetFormatSecOrError(formatId); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = formatSec.Generate(
		rand,
		purpose,
		dst,
	); err != nil {
		err = errors.Wrap(err)
		return
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

	case FormatIdSecEd25519:
		if err = public.SetPurpose(PurposeRepoPubKeyV1); err != nil {
			err = errors.Wrap(err)
			return
		}

		privateBytes := private.GetBytes()
		var privateKey ed25519.PrivateKey

		switch len(privateBytes) {
		case ed25519.SeedSize:
			// TODO emit error
			err = errors.Errorf(
				"private key is just seed, not full go ed25519 private key",
			)
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

		if err = public.SetMarklId(FormatIdPubEd25519, pubKeyBytes); err != nil {
			err = errors.Wrap(err)
			return
		}

	case FormatIdSecAgeX25519:
		if err = public.SetPurpose(PurposeRepoPubKeyV1); err != nil {
			err = errors.Wrap(err)
			return
		}

		// the ed25519 package includes a public key suffix, so we need to
		// reconstruct their version of a private key for a public key value
		privateKey := ed25519.PrivateKey(private.GetBytes())
		pubKeyBytes := privateKey.Public().(ed25519.PublicKey)

		if err = public.SetMarklId(FormatIdPubEd25519, pubKeyBytes); err != nil {
			err = errors.Wrap(err)
			return
		}
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

	case FormatIdSecAgeX25519:
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
