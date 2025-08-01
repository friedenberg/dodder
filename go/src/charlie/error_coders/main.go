package error_coders

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

const (
	pipeTopRight    = "└"
	pipeBottomRight = "┌"
	pipeLeftRight   = "─"
	pipeTeeRight    = "├"
	pipeTopBottom   = "│"
)

type encoder struct{}

var Encoder interfaces.EncoderToWriter[error] = encoder{}

func (encoder encoder) EncodeTo(
	input error,
	writer io.Writer,
) (n int64, err error) {
	bufferedWriter, repool := pool.GetBufferedWriter(writer)
	defer repool()

	if err = encoder.encodeToBufferedWriter(
		input,
		bufferedWriter,
		0,
		false,
	); err != nil {
		err = errors.Wrap(err)
		return
	}
	// TODO calculate N

	if err = bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (encoder encoder) prefixForDepthChild(
	err error,
	depth int,
	lastChild bool,
	continuation bool,
) string {
	if depth == 0 {
		return ""
		// return fmt.Sprintf("%T ", err)
	}

	if continuation {
		return fmt.Sprintf(
			"%s%s ",
			pipeTopBottom,
			strings.Repeat(" ", depth*2),
		)

		// return fmt.Sprintf(
		// 	"%s%s %T ",
		// 	pipeTopBottom,
		// 	strings.Repeat(" ", depth*2),
		// 	err,
		// )
	}

	var firstChar string

	if !lastChild {
		firstChar = pipeTeeRight
	} else {
		firstChar = pipeBottomRight
	}

	return fmt.Sprintf(
		"%s%s ",
		firstChar,
		strings.Repeat(pipeLeftRight, depth*2),
	)
	// return fmt.Sprintf(
	// 	"%s%s %T ",
	// 	firstChar,
	// 	strings.Repeat(pipeLeftRight, depth*2),
	// 	err,
	// )
}

func (encoder encoder) writeOneErrorMessage(
	err error,
	message string,
	bufferedWriter *bufio.Writer,
	depth int,
	lastChild bool,
) {
	prefixWithoutContinuation := encoder.prefixForDepthChild(
		err,
		depth,
		lastChild,
		false,
	)

	prefixWithContinuation := encoder.prefixForDepthChild(
		err,
		depth,
		lastChild,
		true,
	)

	messageReader := bytes.NewBufferString(message)

	var isEOF bool
	var lineIndex int

	for !isEOF {
		line, err := messageReader.ReadBytes('\n')

		line = bytes.TrimSuffix(line, []byte{'\n'})

		isEOF = err == io.EOF

		if len(line) > 0 {
			if lineIndex == 0 {
				fmt.Fprintf(
					bufferedWriter,
					"%s%s\n",
					prefixWithoutContinuation,
					line,
				)
			} else {
				fmt.Fprintf(
					bufferedWriter,
					"%s%s\n",
					prefixWithContinuation,
					line,
				)
			}
		}

		lineIndex++
	}
}

func (encoder encoder) writeOneChildErrorAndFrame(
	err stack_frame.ErrorAndFrame,
	bufferedWriter *bufio.Writer,
	depth int,
	lastChild bool,
) {
	if err.Err != nil {
		encoder.writeOneErrorMessage(
			err,
			fmt.Sprintf("%s\n%s", err.Err, err.Frame),
			bufferedWriter,
			depth+1,
			lastChild,
		)
	} else {
		encoder.writeOneErrorMessage(
			err,
			err.Frame.String(),
			bufferedWriter,
			depth+1,
			lastChild,
		)
	}
}

func (encoder encoder) encodeToBufferedWriter(
	input error,
	bufferedWriter *bufio.Writer,
	depth int,
	lastChild bool,
) (err error) {
	switch input := input.(type) {
	case stack_frame.ErrorsAndFramesGetter:
		children := input.GetErrorsAndFrames()

		lastChildIdx := len(children) - 1

		for idx, child := range slices.Backward(children) {
			encoder.writeOneChildErrorAndFrame(
				child,
				bufferedWriter,
				depth+1,
				idx == lastChildIdx,
			)
		}

	case errors.UnwrapOne:
		encoder.encodeToBufferedWriter(
			input.Unwrap(),
			bufferedWriter,
			depth+1,
			true,
		)

	case errors.UnwrapMany:
		fmt.Fprintf(bufferedWriter, "%T\n", input)

		children := input.Unwrap()
		lastChildIdx := len(children) - 1

		for idx, child := range slices.Backward(children) {
			encoder.encodeToBufferedWriter(
				child,
				bufferedWriter,
				depth+1,
				idx == lastChildIdx,
			)
		}

	case nil:
		encoder.writeOneErrorMessage(
			input,
			"error was nil!",
			bufferedWriter,
			depth,
			lastChild,
		)

		return

	default:
	}

	encoder.writeOneErrorMessage(
		input,
		input.Error(),
		bufferedWriter,
		depth,
		lastChild,
	)

	return
}
