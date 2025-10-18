package xdg_defaults

import "code.linenisgreat.com/dodder/go/src/bravo/env_vars"

const (
	VarXDGHomeOverride = "xdgHomeOverride"
	VarUtilityName     = "utilityName"
)

type DefaultEnvVar struct {
	Name             string
	TemplateDefault  string
	TemplateOverride string
}

var (
	Home = DefaultEnvVar{
		Name:             "HOME",
		TemplateDefault:  "$HOME",
		TemplateOverride: "$HOME",
	}

	Data = DefaultEnvVar{
		Name:             "XDG_DATA_HOME",
		TemplateDefault:  "$HOME/.local/share/$utilityName",
		TemplateOverride: "$xdgHomeOverride/$utilityName/local/share",
	}

	Config = DefaultEnvVar{
		Name:             "XDG_CONFIG_HOME",
		TemplateDefault:  "$HOME/.config/$utilityName",
		TemplateOverride: "$xdgHomeOverride/$utilityName/config",
	}

	State = DefaultEnvVar{
		Name:             "XDG_STATE_HOME",
		TemplateDefault:  "$HOME/.local/state/$utilityName",
		TemplateOverride: "$xdgHomeOverride/$utilityName/local/state",
	}

	Cache = DefaultEnvVar{
		Name:             "XDG_CACHE_HOME",
		TemplateDefault:  "$HOME/.cache/$utilityName",
		TemplateOverride: "$xdgHomeOverride/$utilityName/cache",
	}

	Runtime = DefaultEnvVar{
		Name:             "XDG_RUNTIME_HOME",
		TemplateDefault:  "$HOME/.local/runtime/$utilityName",
		TemplateOverride: "$xdgHomeOverride/$utilityName/local/runtime",
	}
)

func (defaultEnvVar DefaultEnvVar) MakeBaseEnvVar(
	actual string,
) env_vars.DirectoryLayoutBaseEnvVar {
	return env_vars.DirectoryLayoutBaseEnvVar{
		Name:                 defaultEnvVar.Name,
		DefaultValueTemplate: defaultEnvVar.TemplateDefault,
		ActualValue:          actual,
	}
}

func MakeGetenv(
	getenv env_vars.Getenv,
	homeOverride string,
	utilityName string,
) env_vars.Getenv {
	return func(envVarName string) string {
		switch envVarName {
		case VarUtilityName:
			return utilityName

		case VarXDGHomeOverride:
			return homeOverride

		default:
			return getenv(envVarName)
		}
	}
}
