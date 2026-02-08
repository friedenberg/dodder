package ids

import "code.linenisgreat.com/dodder/go/src/charlie/catgut"

type Field struct {
	key, value catgut.String
}

func (f *Field) SetCatgutString(v *catgut.String) (err error) {
	return err
}
