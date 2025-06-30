package xdg

func (xdg *XDG) getInitElements() []xdgInitElement {
	return []xdgInitElement{
		{
			standard:   "$HOME/.local/share",
			overridden: "$HOME/local/share",
			envKey:     "XDG_DATA_HOME",
			out:        &xdg.Data,
		},
		{
			standard:   "$HOME/.config",
			overridden: "$HOME/config",
			envKey:     "XDG_CONFIG_HOME",
			out:        &xdg.Config,
		},
		{
			standard:   "$HOME/.local/state",
			overridden: "$HOME/local/state",
			envKey:     "XDG_STATE_HOME",
			out:        &xdg.State,
		},
		{
			standard:   "$HOME/.cache",
			overridden: "$HOME/cache",
			envKey:     "XDG_CACHE_HOME",
			out:        &xdg.Cache,
		},
		{
			standard:   "$HOME/.local/runtime",
			overridden: "$HOME/local/runtime",
			envKey:     "XDG_RUNTIME_HOME",
			out:        &xdg.Runtime,
		},
	}
}
