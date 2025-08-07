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
