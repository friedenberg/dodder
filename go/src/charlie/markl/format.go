package markl

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

// actual formats
const (
	// TODO maybe switch to type-prefixed
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
	// Ed22519
	makeFormat(FormatPub{
		Id:     FormatIdPubEd25519,
		Verify: Ed25519Verify,
	})

	makeFormat(
		FormatSec{
			Id: FormatIdSecEd25519,

			Generate: Ed25519GeneratePrivateKey,

			PubFormatId:  FormatIdPubEd25519,
			GetPublicKey: Ed25519GetPublicKey,

			SigFormatId: FormatIdSigEd25519,
			Sign:        Ed25519Sign,
		},
	)

	makeFormat(Format{id: FormatIdSigEd25519})

	// AgeX25519
	makeFormat(Format{id: FormatIdPubAgeX25519})
	makeFormat(
		FormatSec{
			Id:           FormatIdSecAgeX25519,
			Generate:     AgeX25519Generate,
			GetIOWrapper: AgeX25519GetIOWrapper,
		},
	)

	// Nonce
	makeFormat(
		FormatSec{
			Id:       FormatIdSecNonce,
			Generate: NonceGenerate,
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

// move to Id
func GetFormatSecOrError(
	formatIdGetter interfaces.MarklFormatGetter,
) (formatSec FormatSec, err error) {
	format := formatIdGetter.GetMarklFormat()

	if format == nil {
		err = errors.Errorf("empty format for getter: %s", formatIdGetter)
		return
	}

	formatId := formatIdGetter.GetMarklFormat().GetMarklFormatId()

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

type FormatId string

func (formatId FormatId) GetMarklFormat() interfaces.MarklFormat {
	format, err := GetFormatOrError(string(formatId))
	errors.PanicIfError(err)
	return format
}

type Format struct {
	id string
}

var _ interfaces.MarklFormat = Format{}

func (format Format) GetMarklFormatId() string {
	return format.id
}

func makeFormat(format interfaces.MarklFormat) {
	if format == nil {
		panic("nil format")
	}

	formatId := format.GetMarklFormatId()

	if formatId == "" {
		panic("empty formatId")
	}

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

	formats[formatId] = format
}
