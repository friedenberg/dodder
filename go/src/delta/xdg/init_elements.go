package xdg

func (x *XDG) getInitElements() []xdgInitElement {
	return []xdgInitElement{
		{
			standard:   "$HOME/.local/share",
			overridden: "$HOME/local/share",
			envKey:     "XDG_DATA_HOME",
			out:        &x.Data,
		},
		{
			standard:   "$HOME/.config",
			overridden: "$HOME/config",
			envKey:     "XDG_CONFIG_HOME",
			out:        &x.Config,
		},
		{
			standard:   "$HOME/.local/state",
			overridden: "$HOME/local/state",
			envKey:     "XDG_STATE_HOME",
			out:        &x.State,
		},
		{
			standard:   "$HOME/.cache",
			overridden: "$HOME/cache",
			envKey:     "XDG_CACHE_HOME",
			out:        &x.Cache,
		},
		{
			standard:   "$HOME/.local/runtime",
			overridden: "$HOME/local/runtime",
			envKey:     "XDG_RUNTIME_HOME",
			out:        &x.Runtime,
		},
	}
}
