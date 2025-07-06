package local_working_copy

func (local *Repo) Reindex() {
	local.Must(local.Lock)
	local.Must(local.config.Reset)
	local.Must(local.GetStore().Reindex)
	local.Must(local.Unlock)
}
