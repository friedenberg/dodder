package env_ui

type OptionsGetter interface {
	GetEnvOptions() Options
}

type Options struct {
	UIFileIsStderr bool
	IgnoreTtyState bool
}
