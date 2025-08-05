//go:build test

package repo_configs

type TestDryRunOnly struct {
	DryRun bool
}

func (config *TestDryRunOnly) IsDryRun() bool {
	return config.DryRun
}

func (config *TestDryRunOnly) SetDryRun(v bool) {
	config.DryRun = v
}
