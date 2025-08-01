package errors

import "fmt"

type Group []error

func (group Group) Error() string {
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

func (group Group) Unwrap() []error {
	return group
}

func (group Group) Len() int {
	return len(group)
}
