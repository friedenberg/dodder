package digests

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func GetDigest(tipe string) (interfaces.Digest, func()) {
	if env, ok := envs[tipe]; ok {
		digest := env.GetDigest()
		return digest, func() { env.PutDigest(digest) }
	} else {
		panic(errors.Errorf("no env registered for digest type: %s", tipe))
	}
}

func PutDigest(digest interfaces.Digest) {
	tipe := digest.GetType()

	if env, ok := envs[tipe]; ok {
		env.PutDigest(digest)
	} else {
		panic(errors.Errorf("no env registered for digest type: %s", tipe))
	}
}
