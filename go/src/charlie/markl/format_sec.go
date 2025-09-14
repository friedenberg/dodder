package markl

import (
	"crypto/rand"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type FuncFormatSecGenerate func(io.Reader) ([]byte, error)

type FormatSec struct {
	id           string
	funcGenerate FuncFormatSecGenerate
}

var _ interfaces.MarklFormat = FormatSec{}

func GetFormatSecOrError(formatId string) (formatSec FormatSec, err error) {
	var format interfaces.MarklFormat
	if format, err = GetFormatOrError(formatId); err != nil {
		err = errors.Wrap(err)
		return
	}

	var ok bool

	if formatSec, ok = format.(FormatSec); !ok {
		err = errors.Errorf(
			"requested format is not FormatSec, but %T:%s",
			formatSec,
			formatId,
		)
		return
	}

	return
}

func makeFormatSec(
	formatId string,
	funcGenerate func(io.Reader) ([]byte, error),
) {
	makeFormat(
		formatId,
		FormatSec{
			id:           formatId,
			funcGenerate: funcGenerate,
		},
	)
}

func (format FormatSec) GetMarklFormatId() string {
	return format.id
}

func (format FormatSec) Generate(
	readerRand io.Reader,
	purpose string,
	dst interfaces.MutableMarklId,
) (err error) {
	if readerRand == nil {
		readerRand = rand.Reader
	}

	var bites []byte

	if bites, err = format.funcGenerate(readerRand); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = dst.SetPurpose(purpose); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = dst.SetMarklId(format.id, bites); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// func (format FormatSec) GetPublicKey(private interfaces.MarklId) (public Id,
// err error) {
// 	marklTypeId := private.GetMarklFormat().GetMarklFormatId()

// 	switch marklTypeId {
// 	default:
// 		err = errors.Errorf(
// 			"unsupported id: %q. Type: %q",
// 			private.StringWithFormat(),
// 			marklTypeId,
// 		)
// 		return

// 	case PurposeRepoPrivateKeyV1:
// 		// legacy
// 		fallthrough

// 	case FormatIdSecEd25519:
// 		if err = public.SetPurpose(PurposeRepoPubKeyV1); err != nil {
// 			err = errors.Wrap(err)
// 			return
// 		}

// 		privateBytes := private.GetBytes()
// 		var privateKey ed25519.PrivateKey

// 		switch len(privateBytes) {
// 		case ed25519.SeedSize:
// 			// TODO emit error
// 			err = errors.Errorf(
// 				"private key is just seed, not full go ed25519 private key",
// 			)
// 			return
// 			privateKey = ed25519.NewKeyFromSeed(privateBytes)

// 		case ed25519.PrivateKeySize:
// 			privateKey = ed25519.PrivateKey(privateBytes)

// 		default:
// 			err = errors.Errorf("unsupported key size: %d", len(privateBytes))
// 			return
// 		}

// 		pubKey := privateKey.Public()
// 		pubKeyBytes := pubKey.(ed25519.PublicKey)

// 		if err = public.SetMarklId(FormatIdPubEd25519, pubKeyBytes); err != nil {
// 			err = errors.Wrap(err)
// 			return
// 		}

// 	case FormatIdSecAgeX25519:
// 		if err = public.SetPurpose(PurposeRepoPubKeyV1); err != nil {
// 			err = errors.Wrap(err)
// 			return
// 		}

// 		// the ed25519 package includes a public key suffix, so we need to
// 		// reconstruct their version of a private key for a public key value
// 		privateKey := ed25519.PrivateKey(private.GetBytes())
// 		pubKeyBytes := privateKey.Public().(ed25519.PublicKey)

// 		if err = public.SetMarklId(FormatIdPubEd25519, pubKeyBytes); err != nil {
// 			err = errors.Wrap(err)
// 			return
// 		}
// 	}

// 	return
// }
