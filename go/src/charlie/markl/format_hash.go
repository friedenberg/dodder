package markl

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"golang.org/x/crypto/blake2b"
)

type FormatHash struct {
	pool interfaces.PoolValue[Hash]
	id   string
	null Id
}

var (
	_ interfaces.MarklFormat = FormatHash{}
	_ interfaces.FormatHash  = FormatHash{}

	formatHashes map[string]FormatHash = map[string]FormatHash{}

	// TODO remove unnecessary references
	FormatHashSha256     FormatHash
	FormatHashBlake2b256 FormatHash
)

func init() {
	FormatHashSha256 = makeFormatHash(
		sha256.New,
		FormatIdHashSha256,
		&FormatHashSha256,
	)

	FormatHashBlake2b256 = makeFormatHash(
		func() hash.Hash {
			hash, _ := blake2b.New256(nil)
			return hash
		},
		FormatIdHashBlake2b256,
		&FormatHashBlake2b256,
	)
}

func makeFormatHash(
	constructor func() hash.Hash,
	id string,
	self *FormatHash,
) FormatHash {
	_, alreadyExists := formats[id]

	if alreadyExists {
		panic(fmt.Sprintf("hash type already registered: %q", id))
	}

	formatHash := FormatHash{
		pool: pool.MakeValue(
			func() Hash {
				return Hash{
					hash:       constructor(),
					formatHash: self,
				}
			},
			func(hash Hash) {
				hash.Reset()
			},
		),
		id: id,
	}

	hash := constructor()
	formatHash.null.format = self
	formatHash.null.allocDataIfNecessary(hash.Size())
	formatHash.null.data = hash.Sum(formatHash.null.data)

	formats[id] = formatHash
	formatHashes[id] = formatHash

	return formatHash
}

func GetFormatHashOrError(
	formatHashId string,
) (formatHash FormatHash, err error) {
	var ok bool
	formatHash, ok = formatHashes[formatHashId]

	if !ok {
		err = errors.Errorf("unknown hash format: %q", formatHashId)
		return
	}

	return
}

func (formatHash FormatHash) GetHash() interfaces.Hash {
	return formatHash.Get()
}

func (formatHash FormatHash) PutHash(hash interfaces.Hash) {
	if correctHashType, ok := hash.(*Hash); ok {
		formatHash.Put(correctHashType)
	} else {
		panic(errors.Errorf("expected type %T but got %T", correctHashType, hash))
	}
}

func (formatHash *FormatHash) Get() *Hash {
	hash := formatHash.pool.Get()
	hash.formatHash = formatHash
	return &hash
}

func (formatHash FormatHash) Put(hash *Hash) {
	errors.PanicIfError(
		MakeErrWrongType(
			formatHash.id,
			hash.GetMarklFormat().GetMarklFormatId(),
		),
	)
	formatHash.pool.Put(*hash)
}

func (formatHash FormatHash) GetMarklFormatId() string {
	return formatHash.id
}

func (formatHash FormatHash) GetSize() int {
	return formatHash.null.GetSize()
}

func (formatHash FormatHash) GetBlobId() (interfaces.MutableMarklId, interfaces.FuncRepool) {
	hash := formatHash.Get()
	defer formatHash.Put(hash)

	return hash.GetMarklId()
}

func (formatHash FormatHash) GetMarklIdForString(
	input string,
) (interfaces.MarklId, interfaces.FuncRepool) {
	hash := formatHash.Get()
	defer formatHash.Put(hash)

	if _, err := io.WriteString(hash, input); err != nil {
		errors.PanicIfError(err)
	}

	return hash.GetMarklId()
}

func (formatHash FormatHash) GetMarklIdForMarklId(
	input interfaces.MarklId,
) (interfaces.MarklId, interfaces.FuncRepool) {
	hash := formatHash.Get()
	defer formatHash.Put(hash)

	if _, err := hash.Write(input.GetBytes()); err != nil {
		errors.PanicIfError(err)
	}

	return hash.GetMarklId()
}

func (formatHash FormatHash) FromStringContent(
	input string,
) interfaces.MarklId {
	id, _ := formatHash.GetMarklIdForString(input)
	return id
}

func (formatHash FormatHash) GetMarklIdFromStringFormat(
	format string,
	args ...any,
) (interfaces.MarklId, interfaces.FuncRepool) {
	hash := formatHash.Get()
	defer formatHash.Put(hash)

	if _, err := fmt.Fprintf(hash, format, args...); err != nil {
		errors.PanicIfError(err)
	}

	return hash.GetMarklId()
}

func (formatHash FormatHash) GetBlobIdForHexString(
	input string,
) (interfaces.MarklId, interfaces.FuncRepool) {
	hash := formatHash.pool.Get()
	defer formatHash.pool.Put(hash)

	id, repool := hash.GetMarklId()

	errors.PanicIfError(SetHexBytes(formatHash.id, id, []byte(input)))

	return id, repool
}
