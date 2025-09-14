package markl

import (
	"crypto/ed25519"
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/zit/go/zit/src/bravo/bech32"
)

// actual formats
const (
	// keep sorted
	FormatIdPubEd25519 = "ed25519_pub"
	FormatIdSecEd25519 = "ed25519_sec"
	FormatIdSigEd25519 = "ed25519_sig"

	FormatIdPubAgeX25519 = "age_x25519_pub"
	FormatIdSecAgeX25519 = "age_x25519_sec"

	FormatIdHashSha256     = "sha256"
	FormatIdHashBlake2b256 = "blake2b256"

	FormatIdSecNonce = "nonce"
)

func init() {
	makeFormat(FormatIdPubEd25519, nil)
	makeFormatSec(
		FormatIdSecEd25519,
		func(rand io.Reader) (bites []byte, err error) {
			if _, bites, err = ed25519.GenerateKey(rand); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		},
	)

	makeFormat(FormatIdSigEd25519, nil)

	makeFormat(FormatIdPubAgeX25519, nil)
	makeFormatSec(
		FormatIdSecAgeX25519,
		func(_ io.Reader) (bites []byte, err error) {
			var ageId age.Identity

			if err = ageId.GenerateIfNecessary(); err != nil {
				err = errors.Wrap(err)
				return
			}

			bech32String := ageId.String()

			if _, bites, err = bech32.Decode(bech32String); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		},
	)

	makeFormatSec(
		FormatIdSecNonce,
		func(rand io.Reader) (bites []byte, err error) {
			bites = make([]byte, 32)

			if _, err = rand.Read(bites); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		},
	)
}

var formats map[string]interfaces.MarklFormat = map[string]interfaces.MarklFormat{}

func GetFormatOrError(formatId string) (interfaces.MarklFormat, error) {
	if formatId == "zit-repo-private_key-v1" {
		formatId = PurposeRepoPrivateKeyV1
	}

	format, ok := formats[formatId]

	if !ok {
		err := errors.Errorf("unknown format: %q", formatId)
		return nil, err
	}

	return format, nil
}

type Format struct {
	id string
}

var _ interfaces.MarklFormat = Format{}

func (format Format) GetMarklFormatId() string {
	return format.id
}

func makeFormat(formatId string, format interfaces.MarklFormat) {
	existing, alreadyExists := formats[formatId]

	if alreadyExists {
		panic(
			fmt.Sprintf(
				"format already registered: %q (%T)",
				formatId,
				existing,
			),
		)
	}

	if format == nil {
		format = Format{
			id: formatId,
		}
	}

	formats[formatId] = format
}
