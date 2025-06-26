package env_repo

type ErrNotInDodderDir struct{}

func (e ErrNotInDodderDir) Error() string {
	return "not in a dodder directory"
}

func (e ErrNotInDodderDir) ShouldShowStackTrace() bool {
	return false
}

func (e ErrNotInDodderDir) Is(target error) (ok bool) {
	_, ok = target.(ErrNotInDodderDir)
	return
}
