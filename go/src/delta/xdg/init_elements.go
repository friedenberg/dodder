package xdg

type xdgInitElement struct {
	standard   string
	overridden string
	envKey     string
	out        *string
}

func (exdg *XDG) getInitElements() []xdgInitElement {
	return []xdgInitElement{
		{
			standard:   "$HOME/.local/share",
			overridden: "$HOME/local/share",
			envKey:     "XDG_DATA_HOME",
			out:        &exdg.Data,
		},
		{
			standard:   "$HOME/.config",
			overridden: "$HOME/config",
			envKey:     "XDG_CONFIG_HOME",
			out:        &exdg.Config,
		},
		{
			standard:   "$HOME/.local/state",
			overridden: "$HOME/local/state",
			envKey:     "XDG_STATE_HOME",
			out:        &exdg.State,
		},
		{
			standard:   "$HOME/.cache",
			overridden: "$HOME/cache",
			envKey:     "XDG_CACHE_HOME",
			out:        &exdg.Cache,
		},
		{
			standard:   "$HOME/.local/runtime",
			overridden: "$HOME/local/runtime",
			envKey:     "XDG_RUNTIME_HOME",
			out:        &exdg.Runtime,
		},
	}
}
