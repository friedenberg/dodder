package flags

import (
	"fmt"
	"strconv"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

// ErrHelp is the error returned if the -help or -h flag is invoked
// but no such flag is defined.
var ErrHelp = errors.New("flag: help requested")

// errParse is returned by Set if a flag's value fails to parse, such as with an
// invalid integer for Int.
// It then gets wrapped through failf to provide more information.
var errParse = errors.New("parse error")

// errRange is returned by Set if a flag's value is out of range.
// It then gets wrapped through failf to provide more information.
var errRange = errors.New("value out of range")

func numError(err error) error {
	ne, ok := err.(*strconv.NumError)
	if !ok {
		return err
	}
	if ne.Err == strconv.ErrSyntax {
		return errParse
	}
	if ne.Err == strconv.ErrRange {
		return errRange
	}
	return err
}

type ErrInvalidValue struct {
	Name     string
	Actual   string
	Expected []string
}

func (err ErrInvalidValue) Error() string {
	var sb strings.Builder

	if err.Name != "" {
		fmt.Fprintf(
			&sb,
			"unsupported value for `-%s`: %q\n",
			err.Name,
			err.Actual,
		)
	} else {
		fmt.Fprintf(&sb, "unsupported value: %q\n", err.Actual)
	}

	fmt.Fprintf(&sb, "supported values:\n")

	for _, value := range err.Expected {
		fmt.Fprintf(&sb, "- %s\n", value)
	}

	return sb.String()
}

func (err ErrInvalidValue) Is(target error) bool {
	_, ok := target.(ErrInvalidValue)
	return ok
}
