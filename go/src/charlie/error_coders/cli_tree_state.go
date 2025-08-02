package error_coders

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
)

type cliTreeState struct {
	bufferedWriter *bufio.Writer
	bytesWritten   int

	stack cliTreeStateStack
	// depth      int
	// childIdx   int
	// childCount int
}

func (state *cliTreeState) encode(
	input error,
) (err error) {
	state.stack.push(nil, input)
	state.encodeStack()

	if err = state.bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO write instead of return string
func (state *cliTreeState) prefixForDepthChild(
	err error,
	continuation bool,
) string {
	depth := state.stack.getDepth()

	if depth == 0 {
		return ""
		// return fmt.Sprintf("%T ", err)
	}

	var count int

	if depth > 0 {
		count = (depth - 1) * 4
	}

	leftPadding := strings.Repeat(" ", count)

	var pipe string

	if state.stack.getLast().isLastChild() {
		pipe = pipeTopRight
	} else {
		pipe = pipeTeeRight
	}

	return fmt.Sprintf(
		"%s%s%s%s ",
		leftPadding,
		pipe,
		pipeLeftRight,
		pipeLeftRight,
	)
}

func (state *cliTreeState) writeOneErrorMessage(
	err error,
	message string,
) {
	prefixWithoutContinuation := state.prefixForDepthChild(
		err,
		false,
	)

	prefixWithContinuation := state.prefixForDepthChild(
		err,
		true,
	)

	messageReader := bytes.NewBufferString(message)

	var isEOF bool
	var lineIndex int
	var bytesWritten int

	for !isEOF {
		line, err := messageReader.ReadBytes('\n')

		line = bytes.TrimSuffix(line, []byte{'\n'})

		isEOF = err == io.EOF

		if len(line) > 0 {
			if lineIndex == 0 {
				bytesWritten, _ = fmt.Fprintf(
					state.bufferedWriter,
					"%s%s\n",
					prefixWithoutContinuation,
					line,
				)
			} else {
				bytesWritten, _ = fmt.Fprintf(
					state.bufferedWriter,
					"%s%s\n",
					prefixWithContinuation,
					line,
				)
			}

			state.bytesWritten += bytesWritten
			bytesWritten = 0
		}

		lineIndex++
	}
}

func (state *cliTreeState) writeOneChildErrorAndFrame(
	err stack_frame.ErrorAndFrame,
) {
	if err.Err != nil {
		state.writeOneErrorMessage(
			err,
			fmt.Sprintf("%s\n%s", err.Err, err.Frame),
		)
	} else {
		state.writeOneErrorMessage(err, err.Frame.String())
	}
}

func (state *cliTreeState) encodeStack() {
	stackItem := state.stack.getLast()
	input := stackItem.child

	switch inputTyped := input.(type) {
	case stack_frame.ErrorsAndFramesGetter:
		state.writeOneErrorMessage(input, input.Error())

		if stackTracer, ok := inputTyped.(stack_frame.ErrorStackTracer); ok &&
			!stackTracer.ShouldShowStackTrace() {
			break
		}

		children := inputTyped.GetErrorsAndFrames()

		childStackItem := state.stack.push(input, nil)
		childStackItem.childCount = len(children)

		var child stack_frame.ErrorAndFrame

		for childStackItem.childIdx, child = range children {
			childStackItem.child = child
			state.writeOneChildErrorAndFrame(child)
		}

	case errors.UnwrapOne:
		child := inputTyped.Unwrap()

		if child == nil {
			state.writeOneErrorMessage(input, input.Error())
			return
		}

		state.writeOneErrorMessage(
			input,
			fmt.Sprintf("%T: %s", input, input.Error()),
		)

		childStackItem := state.stack.push(input, child)
		childStackItem.childCount = 1
		state.encodeStack()
		state.stack.pop()

	case errors.UnwrapMany:
		children := inputTyped.Unwrap()

		state.writeOneErrorMessage(
			input,
			fmt.Sprintf("%T: %s", input, input.Error()),
		)

		childStackItem := state.stack.push(input, nil)
		childStackItem.childCount = len(children)

		for childStackItem.childIdx, childStackItem.child = range children {
			state.encodeStack()
		}

	case nil:
		state.writeOneErrorMessage(
			inputTyped,
			"error was nil!",
		)

		return

	default:
		state.writeOneErrorMessage(
			input,
			input.Error(),
		)
	}
}
