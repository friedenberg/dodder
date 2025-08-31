package markl

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

const (
	// TODO move to ids' builtin types
	// and then add registration
	// keep sorted
	TypeIdEd25519 = "ed25519"
)

func init() {
	makeType(TypeIdEd25519)
}

func GetMarklTypeOrError(typeId string) (interfaces.MarklType, error) {
	tipe, ok := types[typeId]

	if !ok {
		err := errors.Errorf("unknown type: %q", typeId)
		return nil, err
	}

	return tipe, nil
}
