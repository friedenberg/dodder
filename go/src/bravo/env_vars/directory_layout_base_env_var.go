package env_vars

import (
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
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

func (envVar *DirectoryLayoutBaseEnvVar) InitializeXDGEnvVarOrTemplate(
	utilityName string,
	getenv Getenv,
) (err error) {
	if envVar.ActualValue, _ = os.LookupEnv(envVar.Name); envVar.ActualValue != "" {
		envVar.ActualValue = filepath.Join(envVar.ActualValue, utilityName)
	} else {
		if err = envVar.InitializeXDGTemplate(getenv); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (envVar *DirectoryLayoutBaseEnvVar) InitializeXDGTemplate(
	getenv Getenv,
) (err error) {
	envVar.ActualValue = os.Expand(envVar.DefaultValueTemplate, getenv)

	return err
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
