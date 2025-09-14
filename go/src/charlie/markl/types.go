package markl

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

const (
	// TODO move to ids' builtin types
	// and then add registration
	// keep sorted
	TypeIdEd25519Pub = "ed25519_pub"
	TypeIdEd25519Sec = "ed25519_sec"
	TypeIdEd25519Sig = "ed25519_sig"

	TypeIdAgeX25519Pub = "age_x25519_pub"
	TypeIdAgeX25519Sec = "age_x25519_sec"

	TypeIdNonce = "nonce"
)

func init() {
	makeType(TypeIdEd25519Pub)
	makeType(TypeIdEd25519Sec)
	makeType(TypeIdEd25519Sig)

	makeType(TypeIdAgeX25519Pub)
	makeType(TypeIdAgeX25519Sec)

	makeType(TypeIdNonce)
}

func GetMarklTypeOrError(typeId string) (interfaces.MarklFormat, error) {
	if typeId == "zit-repo-private_key-v1" {
		typeId = PurposeRepoPrivateKeyV1
	}

	tipe, ok := types[typeId]

	if !ok {
		err := errors.Errorf("unknown type: %q", typeId)
		return nil, err
	}

	return tipe, nil
}
