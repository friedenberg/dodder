package config_immutable

import (
	"compress/gzip"
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"github.com/DataDog/zstd"
)

type ErrUnsupportedCompression string

func (e ErrUnsupportedCompression) Error() string {
	return fmt.Sprintf("unsupported compression type: %q", string(e))
}

func (e ErrUnsupportedCompression) Is(err error) (ok bool) {
	_, ok = err.(ErrUnsupportedCompression)
	return
}

const (
	CompressionTypeEmpty = CompressionType("")
	CompressionTypeNone  = CompressionType("none")
	CompressionTypeGzip  = CompressionType("gzip")
	CompressionTypeZlib  = CompressionType("zlib")
	CompressionTypeZstd  = CompressionType("zstd")

	CompressionTypeDefault = CompressionTypeZstd
)

type CompressionType string

func (ct *CompressionType) GetBlobCompression() interfaces.BlobCompression {
	return ct
}

func (ct *CompressionType) SetFlagSet(f *flag.FlagSet) {
	f.Var(ct, "compression-type", "")
}

func (ct CompressionType) String() string {
	return string(ct)
}

func (ct *CompressionType) Set(v string) (err error) {
	v1 := CompressionType(strings.TrimSpace(strings.ToLower(v)))

	switch v1 {
	case CompressionTypeGzip,
		CompressionTypeNone,
		CompressionTypeEmpty,
		CompressionTypeZstd,
		CompressionTypeZlib:
		*ct = v1

	default:
		err = ErrUnsupportedCompression(v)
	}

	return
}

func (ct CompressionType) WrapReader(
	r io.Reader,
) (out io.ReadCloser, err error) {
	switch ct {
	case CompressionTypeGzip:
		out, err = gzip.NewReader(r)

	case CompressionTypeZlib:
		out, err = zlib.NewReader(r)

	case CompressionTypeZstd:
		out = zstd.NewReader(r)

	default:
		out = io.NopCloser(r)
	}

	return
}

func (ct CompressionType) WrapWriter(w io.Writer) (io.WriteCloser, error) {
	switch ct {
	case CompressionTypeGzip:
		return gzip.NewWriter(w), nil

	case CompressionTypeZlib:
		return zlib.NewWriter(w), nil

	case CompressionTypeZstd:
		return zstd.NewWriter(w), nil

	default:
		return files.NopWriteCloser{Writer: w}, nil
	}
}
