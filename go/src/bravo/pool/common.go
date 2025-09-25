package pool

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"hash"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/pool_value"
)

var (
	bufioReader   = Make[bufio.Reader](nil, nil)
	bufioWriter   = Make[bufio.Writer](nil, nil)
	byteReaders   = Make[bytes.Reader](nil, nil)
	stringReaders = Make[strings.Reader](nil, nil)
	sha256Hash    = pool_value.Make(
		func() hash.Hash {
			return sha256.New()
		},
		func(hash hash.Hash) {
			hash.Reset()
		},
	)
)

func GetStringReader(
	value string,
) (stringReader *strings.Reader, repool func()) {
	stringReader = stringReaders.Get()
	stringReader.Reset(value)

	repool = func() {
		stringReaders.Put(stringReader)
	}

	return stringReader, repool
}

func GetByteReader(
	value []byte,
) (byteReader *bytes.Reader, repool func()) {
	byteReader = byteReaders.Get()
	byteReader.Reset(value)

	repool = func() {
		byteReaders.Put(byteReader)
	}

	return byteReader, repool
}

func GetSha256Hash() (hash hash.Hash, repool func()) {
	hash = sha256Hash.Get()

	repool = func() {
		sha256Hash.Put(hash)
	}

	return hash, repool
}

func GetBufferedWriter(
	writer io.Writer,
) (bufferedWriter *bufio.Writer, repool func()) {
	bufferedWriter = bufioWriter.Get()
	bufferedWriter.Reset(writer)

	repool = func() {
		bufioWriter.Put(bufferedWriter)
	}

	return bufferedWriter, repool
}

func GetBufferedReader(
	reader io.Reader,
) (bufferedReader *bufio.Reader, repool func()) {
	bufferedReader = bufioReader.Get()
	bufferedReader.Reset(reader)

	repool = func() {
		bufioReader.Put(bufferedReader)
	}

	return bufferedReader, repool
}
