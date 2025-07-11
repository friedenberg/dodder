package interfaces

type ConfigDryRunReader interface {
	IsDryRun() bool
}

type ConfigDryRunWriter interface {
	SetDryRun(bool)
}

type MutableConfigDryRun interface {
	ConfigDryRunReader
	ConfigDryRunWriter
}

type ConfigGetFilters interface {
	GetFilters() map[string]string
}

type MutableStoredConfig interface {
	ConfigGetFilters
}

type MutableConfig interface {
	MutableStoredConfig
	UsePrintTime() bool
	UsePredictableZettelIds() bool
	MutableConfigDryRun
}

type Config interface {
	MutableConfig
	GetTypeStringFromExtension(t string) string
}
