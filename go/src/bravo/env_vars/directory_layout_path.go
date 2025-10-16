package env_vars

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type DirectoryLayoutPath struct {
	envVar   interfaces.DirectoryLayoutBaseEnvVar
	target   string
	fullPath string
}

var _ interfaces.DirectoryLayoutPath = DirectoryLayoutPath{}

func (path DirectoryLayoutPath) GetBaseEnvVar() interfaces.DirectoryLayoutBaseEnvVar {
	return path.envVar
}

func (path DirectoryLayoutPath) GetTarget() string {
	return path.target
}

func (path DirectoryLayoutPath) String() string {
	return path.fullPath
}

func (path DirectoryLayoutPath) GetTemplate() string {
	return filepath.Join(path.envVar.GetBaseEnvVarName(), path.target)
}
