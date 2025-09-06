package values

import (
	"strconv"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type IntSlice []int

var (
	_ interfaces.Stringer  = IntSlice{}
	_ interfaces.FlagValue = &IntSlice{}
)

func (slice IntSlice) String() string {
	stringSlice := make([]string, len(slice))

	for i := range slice {
		stringSlice[i] = strconv.Itoa(slice[i])
	}

	return strings.Join(stringSlice, ",")
}

func (slice *IntSlice) Set(value string) (err error) {
	value = strings.TrimSpace(value)

	elements := strings.Split(value, ",")

	if len(elements) == 0 {
		err = errors.Errorf(
			"invalid format, expected at least one number but got: %q",
			value,
		)
		return
	}

	*slice = make([]int, len(elements))

	for i := range elements {
		var n int

		if n, err = strconv.Atoi(elements[i]); err != nil {
			err = errors.Wrap(err)
			return
		}

		(*slice)[i] = n
	}

	return
}
