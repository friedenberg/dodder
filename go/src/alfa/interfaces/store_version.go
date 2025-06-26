package interfaces

// TODO combine with config_immutable.StoreVersion and make a sealed struct
type StoreVersion interface {
	Stringer
	GetInt() int
}
