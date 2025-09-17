package markl

import (
	"crypto/ed25519"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"golang.org/x/crypto/curve25519"
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
	makeFormat(
		FormatPub{
			Id:     FormatIdPubEd25519,
			Size:   ed25519.PublicKeySize,
			Verify: Ed25519Verify,
		},
	)

	makeFormat(
		FormatSec{
			Id:   FormatIdSecEd25519,
			Size: ed25519.PrivateKeySize,

			Generate: Ed25519GeneratePrivateKey,

			PubFormatId:  FormatIdPubEd25519,
			GetPublicKey: Ed25519GetPublicKey,

			SigFormatId: FormatIdSigEd25519,
			Sign:        Ed25519Sign,
		},
	)

	makeFormat(
		Format{
			Id:   FormatIdSigEd25519,
			Size: ed25519.SignatureSize,
		},
	)

	// AgeX25519
	makeFormat(
		Format{
			Id:   FormatIdPubAgeX25519,
			Size: curve25519.ScalarSize,
		},
	)
	makeFormat(
		FormatSec{
			Id:           FormatIdSecAgeX25519,
			Size:         curve25519.ScalarSize,
			Generate:     AgeX25519Generate,
			GetIOWrapper: AgeX25519GetIOWrapper,
		},
	)

	// Nonce
	makeFormat(
		FormatSec{
			Id:       FormatIdSecNonce,
			Size:     32,
			Generate: NonceGenerate32,
		},
	)
}

var formats map[string]interfaces.MarklFormat = map[string]interfaces.MarklFormat{}

func GetFormatOrError(formatId string) (interfaces.MarklFormat, error) {
	switch formatId {
	case "zit-repo-private_key-v1", "dodder-repo-private_key-v1":
		formatId = FormatIdSecEd25519
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
	Id   string
	Size int
}

var _ interfaces.MarklFormat = Format{}

func (format Format) GetMarklFormatId() string {
	return format.Id
}

func (format Format) GetSize() int {
	return format.Size
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
