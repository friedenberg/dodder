package blob_ids

import (
	"bytes"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func Equals(a, b interfaces.BlobId) bool {
	return bytes.Equal(a.GetBytes(), b.GetBytes())
}

func Clone(src interfaces.BlobId) interfaces.BlobId {
	env := GetEnv(src.GetType())
	dst := env.GetBlobId()
	errors.PanicIfError(dst.SetDigest(src))
	return dst
}

// Creates a human-readable string representation of a digest.
// TODO add type information
func Format(digest interfaces.BlobId) string {
	return fmt.Sprintf("%x", digest.GetBytes())
}
