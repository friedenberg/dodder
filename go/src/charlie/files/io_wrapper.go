package files

import "io"

type NopeIOWrapper struct{}

func (NopeIOWrapper) WrapReader(
	readerIn io.Reader,
) (readerOut io.ReadCloser, err error) {
	return io.NopCloser(readerIn), nil
}

func (NopeIOWrapper) WrapWriter(
	writerIn io.Writer,
) (io.WriteCloser, error) {
	return NopWriteCloser{Writer: writerIn}, nil
}
