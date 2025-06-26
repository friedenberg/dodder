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

func (t *Type) Set(v string) (err error) {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "archive":
		*t = TypeArchive

	case "", "working-copy":
		*t = TypeWorkingCopy

	default:
		err = errors.Wrapf(ErrUnsupportedRepoType{}, "Value: %q", v)
		return
	}

	return
}

func (t Type) String() string {
	switch t {
	case TypeWorkingCopy:
		return "working-copy"

	case TypeArchive:
		return "archive"

	default:
		return fmt.Sprintf("unknown-%d", t)
	}
}

func (t Type) MarshalText() (b []byte, err error) {
	b = []byte(t.String())
	return
}

func (t *Type) UnmarshalText(b []byte) (err error) {
	if err = t.Set(string(b)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
