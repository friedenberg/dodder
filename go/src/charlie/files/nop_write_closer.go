package files

import "io"

type NopWriteCloser struct {
	io.Writer
}

func (_ NopWriteCloser) Close() error {
	return nil
}
