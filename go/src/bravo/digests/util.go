package digests

import (
	"bytes"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func Equals(a, b interfaces.Digest) bool {
	return bytes.Equal(a.GetBytes(), b.GetBytes())
}

// Creates a human-readable string representation of a digest.
// TODO add type information
func Format(digest interfaces.Digest) string {
	return fmt.Sprintf("%x", digest.GetBytes())
}
