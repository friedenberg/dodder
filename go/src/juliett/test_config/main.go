package test_config

type Config struct {
	DryRun bool
}

func (k *Config) IsDryRun() bool {
	return k.DryRun
}

func (k *Config) SetDryRun(v bool) {
	k.DryRun = v
}

func (k *Config) GetFilters() map[string]string {
	return make(map[string]string)
}
