package pool

import "code.linenisgreat.com/dodder/go/src/alfa/errors"

var ErrDoNotRepool = errors.New("do not repool")

func IsDoNotRepool(err error) bool {
	return errors.Is(err, ErrDoNotRepool)
}
