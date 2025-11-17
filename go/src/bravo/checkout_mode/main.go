package checkout_mode

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type (
	Mode            int
	ModeConstructor int
)

type Getter interface {
	GetCheckoutMode() (Mode, error)
}

const (
	None = Mode(iota)
	All  = Mode(^0)

	metadata = Mode(1 << iota)
	blob
	lockfile

	Blob            = ModeConstructor(blob)
	Metadata        = ModeConstructor(metadata)
	MetadataAndBlob = ModeConstructor(metadata | blob)
)

var AvailableModes = []Mode{
	None,
	metadata,
	blob,
	lockfile,
}

func Make(constructors ...ModeConstructor) Mode {
	var mode Mode

	for _, constructor := range constructors {
		mode |= Mode(constructor)
	}

	return mode
}

func (mode Mode) String() string {
	switch {
	case mode == None:
		return "none"

	case mode.IsMetadataOnly():
		return "metadata"

	case mode.IsBlobOnly():
		return "blob"

	case mode.IncludesBlob() && mode.IncludesMetadata():
		return "both"

	default:
		return fmt.Sprintf("invalid(%08b)", mode)
	}
}

func (mode *Mode) Set(value string) (err error) {
	value = strings.ToLower(strings.TrimSpace(value))

	switch value {
	case "":
		*mode = None

	case "metadata":
	case "object":
		*mode = metadata

	case "blob":
		*mode = blob

	case "both":
		*mode = metadata | blob

	default:
		err = errors.ErrorWithStackf(
			"unsupported checkout mode: %s. Available modes: %q",
			value,
			AvailableModes,
		)

		return err
	}

	return err
}

func (mode Mode) IsMetadataOnly() bool {
	return mode == metadata
}

func (mode Mode) IsBlobOnly() bool {
	return mode == blob
}

func (mode Mode) IncludesBlob() bool {
	return mode&blob != 0
}

func (mode Mode) IncludesMetadata() bool {
	return mode&metadata != 0
}

func (mode Mode) IncludesLockfile() bool {
	return mode&lockfile != 0
}

func (mode Mode) IsBlobRecognized() bool {
	return false
}
