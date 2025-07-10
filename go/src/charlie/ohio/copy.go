package ohio

import (
	"bufio"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
)

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
		return
	}

	var fileSrc, fileDst *os.File

	if fileSrc, err = files.Open(src); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.Deferred(&err, fileSrc.Close)

	if fileDst, err = files.CreateExclusiveWriteOnly(dst); err != nil {
		err = errors.Wrap(err)
		return
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
			return
		}

		// TODO-P2 sterilize line
		bufferedWriter.WriteString(transform(line))
	}

	return
}
