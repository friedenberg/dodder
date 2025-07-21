package sha

import "code.linenisgreat.com/dodder/go/src/alfa/interfaces"

type Hasher struct{}

func (hasher Hasher) MakeWriteDigester() interfaces.WriteDigester {
	return nil
}
