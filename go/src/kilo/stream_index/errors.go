package stream_index

import "code.linenisgreat.com/dodder/go/src/alfa/errors"

var errConcurrentPageAccess = errors.New("concurrent page access")

func MakeErrConcurrentPageAccess() error {
	return errors.WrapSkip(2, errConcurrentPageAccess)
}
