package markl

import "code.linenisgreat.com/dodder/go/src/alfa/interfaces"

type IdWithFormat struct {
	Format string
	Id
}

var (
	_ interfaces.MarklIdWithFormat        = IdWithFormat{}
	_ interfaces.MutableMarklIdWithFormat = &IdWithFormat{}
)

func (id *IdWithFormat) Reset() {
	id.Id.Reset()
	id.Format = ""
}

func (id *IdWithFormat) ResetWith(src IdWithFormat) {
	id.Id.ResetWith(src.Id)
	id.Format = src.Format
}

func (id IdWithFormat) GetFormat() string {
	return id.Format
}

func (id *IdWithFormat) SetFormat(value string) error {
	id.Format = value
	return nil
}
