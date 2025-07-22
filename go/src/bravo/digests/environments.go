package digests

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

var (
	envsLock sync.Mutex
	envs     = make(map[string]interfaces.EnvDigest)
)

func RegisterEnv(
	env interfaces.EnvDigest,
) struct{} {
	envsLock.Lock()
	defer envsLock.Unlock()

	tipe := env.GetType()

	if existing, ok := envs[tipe]; ok {
		panic(errors.Errorf("digest env already registered with %T", existing))
	}

	envs[tipe] = env

	return struct{}{}
}

func GetEnv(tipe string) interfaces.EnvDigest {
	if env, ok := envs[tipe]; ok {
		return env
	} else {
		panic(errors.Errorf("no env registered for digest type: %s", tipe))
	}
}
