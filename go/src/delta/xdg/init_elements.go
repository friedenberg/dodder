package xdg

import (
	"code.linenisgreat.com/dodder/go/src/bravo/env_vars"
)

type initElement struct {
	defawlt DefaultEnvVar
	actual  *env_vars.DirectoryLayoutBaseEnvVar
}

func (exdg *XDG) getInitElements() []initElement {
	return []initElement{
		{
			defawlt: DefaultData,
			actual:  &exdg.Data,
		},
		{
			defawlt: DefaultConfig,
			actual:  &exdg.Config,
		},
		{
			defawlt: DefaultState,
			actual:  &exdg.State,
		},
		{
			defawlt: DefaultCache,
			actual:  &exdg.Cache,
		},
		{
			defawlt: DefaultRuntime,
			actual:  &exdg.Runtime,
		},
	}
}
