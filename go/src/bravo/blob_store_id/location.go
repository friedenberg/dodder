package blob_store_id

import "code.linenisgreat.com/dodder/go/src/alfa/errors"

type (
	LocationType interface {
		LocationTypeGetter
		location()
		GetPrefix() rune
	}

	LocationTypeGetter interface {
		GetLocationType() LocationType
	}

	//go:generate stringer -type=location
	location int
)

var (
	_ LocationTypeGetter = location(0)
	_ LocationType       = location(0)
)

func (location) location() {}

func (location location) GetLocationType() LocationType { return location }

func (location *location) SetPrefix(firstChar rune) (err error) {
	switch firstChar {
	case '/':
		*location = LocationTypeXDGSystem

	case '~':
		*location = LocationTypeXDGUser

	case '.':
		*location = LocationTypeOverride

	case '_':
		*location = LocationTypeRemote

	default:
		err = errors.Errorf(
			"unsupported rune for location type: %q",
			string(firstChar),
		)

		return err
	}

	return err
}

func (location location) GetPrefix() rune {
	switch location {
	case LocationTypeXDGSystem:
		return '/'

	case LocationTypeXDGUser:
		return '~'

	case LocationTypeOverride:
		return '.'

	case LocationTypeRemote:
		return '_'

	default:
		panic(errors.Errorf("unsupported location type: %q", location))
	}
}

const (
	LocationTypeUnknown = location(iota)
	LocationTypeXDGSystem
	LocationTypeXDGUser
	LocationTypeOverride
	LocationTypeRemote
)
