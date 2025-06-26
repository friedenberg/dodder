package object_metadata

var Lessor lessor

type lessor struct{}

func (lessor) Less(a, b *Metadata) bool {
	return a.Tai.Less(b.Tai)
}
