package pool

import (
	"bufio"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

var (
	bufioReader interfaces.Pool[bufio.Reader, *bufio.Reader]
	bufioWriter interfaces.Pool[bufio.Writer, *bufio.Writer]
)

func init() {
	bufioReader = MakePool[bufio.Reader](nil, nil)
	bufioWriter = MakePool[bufio.Writer](nil, nil)
}

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
