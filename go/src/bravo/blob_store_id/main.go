package blob_store_id

import (
	"encoding"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type Id struct {
	location LocationType
	id       string
}

var (
	_ interfaces.Stringer      = Id{}
	_ interfaces.Setter        = &Id{}
	_ encoding.TextMarshaler   = Id{}
	_ encoding.TextUnmarshaler = &Id{}
)

func Make(id string) Id {
	return Id{
		location: LocationTypeXDG,
		id:       id,
	}
}

func MakeWithLocation(id string, location LocationTypeGetter) Id {
	return Id{
		location: location.GetLocationType(),
		id:       id,
	}
}

func (id Id) IsEmpty() bool {
	return id.id == ""
}

func (id Id) String() string {
	if id.id == "" {
		return ""
	}

	switch id.location {
	case LocationTypeOverride:
		return fmt.Sprintf(".%s", id.id)

	case LocationTypeXDG:
		return fmt.Sprintf("/%s", id.id)

	case LocationTypeRemote:
		return fmt.Sprintf("_%s", id.id)

	default:
		panic(
			fmt.Sprintf("unknown location for blob store id: %q", id.location),
		)
	}
}

func (id *Id) Set(value string) (err error) {
	var firstChar byte
	firstChar, id.id = value[0], value[1:]

	switch firstChar {
	case '/':
		id.location = LocationTypeXDG

	case '.':
		id.location = LocationTypeOverride

	case '_':
		id.location = LocationTypeRemote

	default:
		err = errors.Errorf(
			"unsupported first char for blob_store_id: %q",
			string(firstChar),
		)

		return err
	}

	return err
}

func (id Id) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *Id) UnmarshalText(bites []byte) (err error) {
	if err = id.Set(string(bites)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
