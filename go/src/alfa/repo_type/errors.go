package repo_type

import "fmt"

type ErrUnsupportedRepoType struct {
	Expected, Actual Type
}

func (err ErrUnsupportedRepoType) Error() string {
	return fmt.Sprintf(
		"%q repos are not supported, only %q",
		err.Actual,
		err.Expected,
	)
}

func (err ErrUnsupportedRepoType) Is(target error) bool {
	_, ok := target.(ErrUnsupportedRepoType)
	return ok
}
