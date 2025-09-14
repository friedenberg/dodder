package markl

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

// actual formats
const (
	// keep sorted
	FormatIdEd25519Pub = "ed25519_pub"
	FormatIdEd25519Sec = "ed25519_sec"
	FormatIdEd25519Sig = "ed25519_sig"

	FormatIdAgeX25519Pub = "age_x25519_pub"
	FormatIdAgeX25519Sec = "age_x25519_sec"

	FormatIdNonce = "nonce"
)

func init() {
	makeFormat(FormatIdEd25519Pub)
	makeFormat(FormatIdEd25519Sec)
	makeFormat(FormatIdEd25519Sig)

	makeFormat(FormatIdAgeX25519Pub)
	makeFormat(FormatIdAgeX25519Sec)

	makeFormat(FormatIdNonce)
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

type format struct {
	id string
}

var _ interfaces.MarklFormat = format{}

func (format format) GetMarklFormatId() string {
	return format.id
}

func makeFormat(formatId string) {
	_, alreadyExists := formats[formatId]

	if alreadyExists {
		panic(fmt.Sprintf("hash type already registered: %q", formatId))
	}

	formats[formatId] = format{
		id: formatId,
	}
}
