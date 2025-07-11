package sha

import (
	"bytes"
	"crypto/sha256"
	"hash"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var (
	poolHash256 interfaces.PoolValue[hash.Hash]
	poolSha     interfaces.Pool[Sha, *Sha]
	poolWriter  interfaces.Pool[writer, *writer]
)

func init() {
	poolHash256 = pool.MakeValue(
		func() hash.Hash {
			return sha256.New()
		},
		func(h hash.Hash) {
			h.Reset()
		},
	)

	poolSha = pool.MakePool(
		nil,
		func(sh *Sha) {
			sh.Reset()
		},
	)

	poolWriter = pool.MakePool(
		nil,
		func(writer *writer) {
			writer.Reset(nil)
		},
	)
}

func GetPool() interfaces.Pool[Sha, *Sha] {
	return poolSha
}

func GetPoolWriter() interfaces.Pool[writer, *writer] {
	return poolWriter
}

var Resetter resetter

type resetter struct{}

func (resetter) Reset(s *Sha) {
	s.Reset()
}

func (resetter) ResetWith(a, b *Sha) {
	a.ResetWith(b)
}

var Lessor lessor

type lessor struct{}

func (lessor) Less(a, b *Sha) bool {
	return bytes.Compare(a.GetShaBytes(), b.GetShaBytes()) == -1
}

var Equaler equaler

type equaler struct{}

func (equaler) Equals(a, b *Sha) bool {
	return bytes.Equal(a.GetShaBytes(), b.GetShaBytes())
}
