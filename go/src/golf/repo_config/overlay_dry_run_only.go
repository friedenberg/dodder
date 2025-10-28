//go:build test

package repo_config

type DryRunOnly struct {
	DryRun bool
}

func (config *DryRunOnly) IsDryRun() bool {
	return config.DryRun
}

func (config *DryRunOnly) SetDryRun(v bool) {
	config.DryRun = v
}
