package repo

import (
	"code.linenisgreat.com/dodder/go/src/golf/config_immutable_io"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/kilo/env_workspace"
)

type LocalRepo interface {
	Repo
	GetEnvRepo() env_repo.Env // TODO rename to GetEnvRepo
	GetImmutableConfigPrivate() config_immutable_io.ConfigLoadedPrivate
	Lock() error
	Unlock() error
}

type LocalWorkingCopy interface {
	WorkingCopy
	LocalRepo
	GetEnvWorkspace() env_workspace.Env
}
