package catgut

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var (
	p     interfaces.Pool[String, *String]
	ponce sync.Once
)

func init() {
}

func GetPool() interfaces.Pool[String, *String] {
	ponce.Do(
		func() {
			p = pool.Make[String, *String](
				nil,
				func(v *String) {
					v.Reset()
				},
			)
		},
	)

	return p
}
