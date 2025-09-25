package ohio

import (
	"bufio"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
)

func CopyBuffered(dst io.Writer, src io.Reader) (written int64, err error) {
	bufferedReader, repoolBufferedReader := pool.GetBufferedReader(src)
	defer repoolBufferedReader()

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(dst)
	defer repoolBufferedWriter()
	defer errors.DeferredFlusher(&err, bufferedWriter)

	if written, err = io.Copy(bufferedWriter, bufferedReader); err != nil {
		err = errors.Wrap(err)
		return written, err
	}

	return written, err
}

func CopyFileLines(
	src, dst string,
) (err error) {
	return CopyFileWithTransform(
		src,
		dst,
		'\n',
		func(value string) string {
			return value
		},
	)
}

func CopyFileWithTransform(
	src, dst string,
	delim byte,
	transform func(string) string,
) (err error) {
	if src == "" {
		return err
	}

	var fileSrc, fileDst *os.File

	if fileSrc, err = files.Open(src); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.Deferred(&err, fileSrc.Close)

	if fileDst, err = files.CreateExclusiveWriteOnly(dst); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.Deferred(&err, fileDst.Close)

	bufferedReader := bufio.NewReader(fileSrc)
	bufferedWriter := bufio.NewWriter(fileDst)

	defer errors.Deferred(&err, bufferedWriter.Flush)

	isEOF := false

	for !isEOF {
		var line string
		line, err = bufferedReader.ReadString(delim)

		if err == io.EOF {
			isEOF = true
			err = nil
		} else if err != nil {
			err = errors.Wrap(err)
			return err
		}

		// TODO-P2 sterilize line
		bufferedWriter.WriteString(transform(line))
	}

	return err
}
