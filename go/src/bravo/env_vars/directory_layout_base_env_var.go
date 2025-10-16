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

func (envVar DirectoryLayoutBaseEnvVar) GetBaseEnvVar() string {
	return envVar.Name
}

func (envVar DirectoryLayoutBaseEnvVar) String() string {
	return envVar.ActualValue
}

func (envVar DirectoryLayoutBaseEnvVar) GetBase() string {
	return envVar.ActualValue
}

func (envVar DirectoryLayoutBaseEnvVar) MakePath(
	targets ...string,
) interfaces.DirectoryLayoutPath {
	target := filepath.Join(targets...)

	return DirectoryLayoutPath{
		envVar:   envVar,
		target:   target,
		fullPath: filepath.Join(envVar.GetBase(), target),
	}
}
