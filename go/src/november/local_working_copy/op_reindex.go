package local_working_copy

func (localWorkingCopy *Repo) Reindex() {
	localWorkingCopy.Must(localWorkingCopy.Lock)
	localWorkingCopy.Must(localWorkingCopy.config.Reset)
	localWorkingCopy.Must(localWorkingCopy.GetStore().Reindex)
	localWorkingCopy.Must(localWorkingCopy.Unlock)
}
