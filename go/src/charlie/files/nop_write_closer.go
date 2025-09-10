package files

import "io"

type nopWriteCloser struct {
	io.Writer
}

func NopWriteCloser(writer io.Writer) nopWriteCloser {
	return nopWriteCloser{Writer: writer}
}

func (nopWriteCloser) Close() error {
	return nil
}
