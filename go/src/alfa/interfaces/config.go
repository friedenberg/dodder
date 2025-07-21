package interfaces

type (
	ConfigDryRunGetter interface {
		IsDryRun() bool
	}

	ConfigDryRunSetter interface {
		SetDryRun(bool)
	}

	MutableConfigDryRun interface {
		ConfigDryRunGetter
		ConfigDryRunSetter
	}

	Config interface {
		UsePrintTime() bool
		UsePredictableZettelIds() bool
		GetTypeStringFromExtension(t string) string
		GetTypeExtension(string) string
		ConfigDryRunGetter
	}

	MutableConfig interface {
		Config
		MutableConfigDryRun
	}
)
