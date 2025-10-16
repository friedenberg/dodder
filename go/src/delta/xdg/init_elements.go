package xdg

type xdgInitElement struct {
	standard DefaultEnvVar
	out      *string
}

func (exdg *XDG) getInitElements() []xdgInitElement {
	return []xdgInitElement{
		{
			standard: DefaultData,
			out:      &exdg.Data.ActualValue,
		},
		{
			standard: DefaultConfig,
			out:      &exdg.Config.ActualValue,
		},
		{
			standard: DefaultState,
			out:      &exdg.State.ActualValue,
		},
		{
			standard: DefaultCache,
			out:      &exdg.Cache.ActualValue,
		},
		{
			standard: DefaultRuntime,
			out:      &exdg.Runtime.ActualValue,
		},
	}
}
