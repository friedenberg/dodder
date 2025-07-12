package genesis_configs

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type ErrFutureStoreVersion struct {
	interfaces.StoreVersion
}

func (err ErrFutureStoreVersion) Error() string {
	return fmt.Sprintf(
		strings.Join(
			[]string{
				"store version is from the future: %q",
				"This means that this installation of dodder is likely out of date.",
			},
			". ",
		),
		err.StoreVersion,
	)
}

func (ErrFutureStoreVersion) Is(target error) bool {
	_, ok := target.(ErrFutureStoreVersion)
	return ok
}
