package object_inventory_format

type Options struct {
	Tai           bool
	ExcludeMutter bool
	Verzeichnisse bool
	PrintFinalSha bool
}

func (o Options) SansVerzeichnisse() Options {
	o.Verzeichnisse = false
	return o
}
