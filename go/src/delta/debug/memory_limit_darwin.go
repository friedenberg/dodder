package debug

import "code.linenisgreat.com/dodder/go/src/alfa/errors"

func getMemoryLimit() (uint64, error) {
  return 0, errors.New("memory limit not supported")
}
