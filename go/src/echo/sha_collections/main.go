package sha_collections

import (
	"crypto/sha256"
	"io"
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

type (
	Set        = interfaces.SetLike[*sha.Sha]
	MutableSet = interfaces.MutableSetLike[*sha.Sha]
)

func init() {
	collections_value.RegisterGobValue[*sha.Sha](nil)
}

func MakeMutableSet(es ...*sha.Sha) (s MutableSet) {
	return collections_value.MakeMutableValueSet(nil, es...)
}

func ShaFromSet(s interfaces.SetLike[sha.Sha]) interfaces.Digest {
	hash := sha256.New()

	elements := make([]sha.Sha, 0, s.Len())

	s.Each(
		func(s sha.Sha) (err error) {
			elements = append(elements, s)
			return
		},
	)

	sort.Slice(
		elements,
		func(i, j int) bool { return elements[i].String() < elements[j].String() },
	)

	for _, e := range elements {
		if _, err := io.WriteString(hash, e.String()); err != nil {
			errors.PanicIfError(errors.Wrap(err))
		}
	}

	return sha.FromHash(hash)
}
