package pool

import (
	"bufio"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

var (
	bufioReader   = MakePool[bufio.Reader](nil, nil)
	bufioWriter   = MakePool[bufio.Writer](nil, nil)
	stringReaders = MakePool[strings.Reader](nil, nil)
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
