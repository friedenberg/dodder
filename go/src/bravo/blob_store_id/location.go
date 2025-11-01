package blob_store_id

type LocationType interface {
	location()
}

//go:generate stringer -type=location
type location int

func (location) location() {}

const (
	LocationTypeUnknown = location(iota)
	LocationTypeXDG
	LocationTypeOverride
	LocationTypeRemote
)
