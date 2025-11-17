package ids

type TypedBlob[T any] struct {
	Type Type
	Blob T
}
