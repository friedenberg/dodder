package sha

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

// TODO replace with digests.GetDigest
func GetPool() interfaces.Pool[Sha, *Sha] {
	return poolSha
}
