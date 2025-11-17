package tag_paths

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
)

var p interfaces.Pool[Path, *Path]

func init() {
	p = pool.Make(
		func() *Path {
			return &Path{}
		},
		func(p *Path) {
			for _, s := range *p {
				s.Reset()
			}

			*p = (*p)[:0]
		},
	)
}

func GetPool() interfaces.Pool[Path, *Path] {
	return p
}
