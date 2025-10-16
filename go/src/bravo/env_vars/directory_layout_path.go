package env_vars

import "code.linenisgreat.com/dodder/go/src/alfa/interfaces"

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
