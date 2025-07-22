package sha

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var poolSha interfaces.Pool[Sha, *Sha]

func init() {
	poolSha = pool.MakePool(
		nil,
		func(sh *Sha) {
			sh.Reset()
		},
	)
}

// TODO replace with digests.GetDigest
func GetPool() interfaces.Pool[Sha, *Sha] {
	return poolSha
}
