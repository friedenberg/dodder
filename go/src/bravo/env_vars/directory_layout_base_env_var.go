package env_vars

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type DirectoryLayoutBaseEnvVar struct {
	Name                 string
	DefaultValueTemplate string
	ActualValue          string
}

var _ interfaces.DirectoryLayoutBaseEnvVar = DirectoryLayoutBaseEnvVar{}

func (envVar DirectoryLayoutBaseEnvVar) GetBaseEnvVarName() string {
	return envVar.Name
}

func (envVar DirectoryLayoutBaseEnvVar) String() string {
	return envVar.ActualValue
}

func (envVar DirectoryLayoutBaseEnvVar) GetBaseEnvVarValue() string {
	return envVar.ActualValue
}

func (envVar DirectoryLayoutBaseEnvVar) MakePath(
	targets ...string,
) interfaces.DirectoryLayoutPath {
	target := filepath.Join(targets...)

	return DirectoryLayoutPath{
		envVar:   envVar,
		target:   target,
		fullPath: filepath.Join(envVar.GetBaseEnvVarValue(), target),
	}
}
