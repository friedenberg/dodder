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
		id:         FormatIdPubEd25519,
		funcVerify: Ed25519Verify,
	})

	makeFormat(
		FormatSec{
			id: FormatIdSecEd25519,

			funcGenerate: Ed25519GeneratePrivateKey,

			pubkeyFormatId:   FormatIdPubEd25519,
			funcGetPublicKey: Ed25519GetPublicKey,

			sigFormatId: FormatIdSigEd25519,
			funcSign:    Ed25519Sign,
		},
	)

	makeFormat(Format{id: FormatIdSigEd25519})

	// AgeX25519
	makeFormat(Format{id: FormatIdPubAgeX25519})
	makeFormat(
		FormatSec{
			id:               FormatIdSecAgeX25519,
			funcGenerate:     AgeX25519Generate,
			funcGetIOWrapper: AgeX25519GetIOWrapper,
		},
	)

	// Nonce
	makeFormat(
		FormatSec{
			id:           FormatIdSecNonce,
			funcGenerate: NonceGenerate,
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

// move to Id
func GetFormatPubOrError(
	formatIdGetter interfaces.MarklFormatGetter,
) (formatPub FormatPub, err error) {
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

	if formatPub, ok = format.(FormatPub); !ok {
		err = errors.Errorf(
			"requested format is not FormatPub, but %T:%s",
			formatPub,
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
