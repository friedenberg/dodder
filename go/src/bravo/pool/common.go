package pool

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"hash"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

var (
	bufioReader   = MakePool[bufio.Reader](nil, nil)
	bufioWriter   = MakePool[bufio.Writer](nil, nil)
	byteReaders   = MakePool[bytes.Reader](nil, nil)
	stringReaders = MakePool[strings.Reader](nil, nil)
	sha256Hash    = MakeValue(
		func() hash.Hash {
			return sha256.New()
		},
		func(hash hash.Hash) {
			hash.Reset()
		},
	)
)

func GetBufioReader() interfaces.Pool[bufio.Reader, *bufio.Reader] {
	return bufioReader
}

func GetBufioWriter() interfaces.Pool[bufio.Writer, *bufio.Writer] {
	return bufioWriter
}

func FlushBufioWriterAndPut(err *error, writer *bufio.Writer) {
	errors.DeferredFlusher(err, writer)
	GetBufioWriter().Put(writer)
}

func GetStringReader(
	value string,
) (stringReader *strings.Reader, repool func()) {
	stringReader = stringReaders.Get()
	stringReader.Reset(value)

	repool = func() {
		stringReaders.Put(stringReader)
	}

	return
}

func GetByteReader(
	value []byte,
) (byteReader *bytes.Reader, repool func()) {
	byteReader = byteReaders.Get()
	byteReader.Reset(value)

	repool = func() {
		byteReaders.Put(byteReader)
	}

	return
}

func GetSha256Hash() (hash hash.Hash, repool func()) {
	hash = sha256Hash.Get()

	repool = func() {
		sha256Hash.Put(hash)
	}

	return
}

func GetBufferedWriter(
	writer io.Writer,
) (bufferedWriter *bufio.Writer, repool func()) {
	bufferedWriter = bufioWriter.Get()
	bufferedWriter.Reset(writer)

	repool = func() {
		bufioWriter.Put(bufferedWriter)
	}

	return
}

func GetBufferedReader(
	reader io.Reader,
) (bufferedReader *bufio.Reader, repool func()) {
	bufferedReader = bufioReader.Get()
	bufferedReader.Reset(reader)

	repool = func() {
		bufioReader.Put(bufferedReader)
	}

	return
}
