package errors

import "fmt"

type group []error

func (group group) Error() string {
	count := group.Len()

	switch count {
	case 0:
		panic("empty error group")

	case 1:
		return group[0].Error()

	default:
		return fmt.Sprintf("%d errors in group", group.Len())
	}
}

func (group group) Unwrap() []error {
	return group
}

func (group group) Len() int {
	return len(group)
}
