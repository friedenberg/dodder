package digests

import (
	"bytes"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func DigesterEquals(a, b interfaces.Digester) bool {
	return DigestEquals(a.GetDigest(), b.GetDigest())
}

func DigestEquals(a, b interfaces.Digest) bool {
	return bytes.Equal(a.GetBytes(), b.GetBytes())
}

func FormatDigester(digester interfaces.Digester) string {
	return FormatDigest(digester.GetDigest())
}

// Creates a human-readable string representation of a digest.
// TODO add type information
func FormatDigest(digest interfaces.Digest) string {
	return fmt.Sprintf("%x", digest.GetBytes())
}
