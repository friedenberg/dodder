package repo_type

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type Type int

const (
	TypeUnknown = Type(iota)
	TypeWorkingCopy
	TypeArchive
)

func (tipe *Type) Set(v string) (err error) {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "archive":
		*tipe = TypeArchive

	case "", "working-copy":
		*tipe = TypeWorkingCopy

	default:
		err = errors.Wrapf(ErrUnsupportedRepoType{}, "Value: %q", v)
		return
	}

	return
}

func (tipe Type) String() string {
	switch tipe {
	case TypeWorkingCopy:
		return "working-copy"

	case TypeArchive:
		return "archive"

	default:
		return fmt.Sprintf("unknown-%d", tipe)
	}
}

func (tipe Type) MarshalText() (b []byte, err error) {
	b = []byte(tipe.String())
	return
}

func (tipe *Type) UnmarshalText(b []byte) (err error) {
	if err = tipe.Set(string(b)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
