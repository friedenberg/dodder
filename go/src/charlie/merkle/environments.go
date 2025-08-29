package merkle

import (
	"fmt"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

// TODO remove

var (
	envsLock sync.Mutex
	envs     = make(map[string]interfaces.EnvBlobId)
)

func RegisterEnv(
	env interfaces.EnvBlobId,
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

func GetEnv(tipe string) interfaces.EnvBlobId {
	if env, ok := envs[tipe]; ok {
		return env
	} else {
		panic(fmt.Sprintf("no env registered for digest type: %s", tipe))
	}
}
