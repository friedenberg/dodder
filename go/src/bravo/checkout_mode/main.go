package checkout_mode

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type Mode int

type Getter interface {
	GetCheckoutMode() (Mode, error)
}

const (
	None = Mode(iota)
	MetadataOnly
	MetadataAndBlob
	BlobOnly

	// TODO remove?
	BlobRecognized // should never be set via flags
)

var AvailableModes = []Mode{
	None,
	MetadataOnly,
	MetadataAndBlob,
	BlobOnly,
}

func (m Mode) String() string {
	switch m {
	case None:
		return "none"

	case MetadataOnly:
		return "metadata"

	case BlobOnly:
		return "blob"

	case MetadataAndBlob:
		return "both"

	default:
		return "unknown"
	}
}

func (m *Mode) Set(v string) (err error) {
	v = strings.ToLower(strings.TrimSpace(v))

	switch v {
	case "":
		*m = None

	case "metadata":
	case "object":
		*m = MetadataOnly

	case "blob":
		*m = BlobOnly

	case "both":
		*m = MetadataAndBlob

	default:
		err = errors.ErrorWithStackf(
			"unsupported checkout mode: %s. Available modes: %q",
			v,
			AvailableModes,
		)

		return err
	}

	return err
}

func (m Mode) IncludesBlob() bool {
	switch m {
	case MetadataAndBlob, BlobOnly:
		return true

	default:
		return false
	}
}

func (m Mode) IncludesMetadata() bool {
	switch m {
	case MetadataAndBlob, MetadataOnly:
		return true

	default:
		return false
	}
}
