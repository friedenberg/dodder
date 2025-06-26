package ohio

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

func BufferedWriter(writer io.Writer) *bufio.Writer {
	bufferedWriter := pool.GetBufioWriter().Get()
	bufferedWriter.Reset(writer)
	return bufferedWriter
}

func BufferedReader(reader io.Reader) *bufio.Reader {
	bufferedReader := pool.GetBufioReader().Get()
	bufferedReader.Reset(reader)
	return bufferedReader
}
