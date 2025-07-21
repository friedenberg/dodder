package sha

import "code.linenisgreat.com/dodder/go/src/alfa/interfaces"

type Env struct{}

func (env Env) MakeWriteDigester() interfaces.WriteDigester {
	return MakeWriter(nil)
}

func (env Env) MakeReadDigester() interfaces.ReadDigester {
	return MakeReadCloser(nil)
}
