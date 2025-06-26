package env_repo

type ErrNotInZitDir struct{}

func (e ErrNotInZitDir) Error() string {
	return "not in a zit directory"
}

func (e ErrNotInZitDir) ShouldShowStackTrace() bool {
	return false
}

func (e ErrNotInZitDir) Is(target error) (ok bool) {
	_, ok = target.(ErrNotInZitDir)
	return
}
