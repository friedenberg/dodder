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
		MutableConfig
		GetTypeStringFromExtension(t string) string
	}

	MutableConfig interface {
		UsePrintTime() bool
		UsePredictableZettelIds() bool
		MutableConfigDryRun
	}
)
