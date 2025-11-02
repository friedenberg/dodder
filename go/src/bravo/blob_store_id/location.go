package blob_store_id

type (
	LocationType interface {
		LocationTypeGetter
		location()
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

const (
	LocationTypeUnknown = location(iota)
	LocationTypeXDG
	LocationTypeOverride
	LocationTypeRemote
)
