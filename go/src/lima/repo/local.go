package repo

import (
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/kilo/env_workspace"
)

type LocalRepo interface {
	Repo
	GetEnvRepo() env_repo.Env // TODO rename to GetEnvRepo
	GetImmutableConfigPrivate() genesis_configs.TypedConfigPrivate
	Lock() error
	Unlock() error
}

type LocalWorkingCopy interface {
	WorkingCopy
	LocalRepo
	GetEnvWorkspace() env_workspace.Env
}
