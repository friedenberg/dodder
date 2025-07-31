package stack_frame

func MakeErrorAndFrames(capInt int) *ErrorsAndFrames {
	output := make(ErrorsAndFrames, 0, capInt)
	errorAndFrames := &output
	return errorAndFrames
}

type ErrorsAndFrames []ErrorAndFrame

func (err *ErrorsAndFrames) AppendFrames(frames []Frame) {
	for _, frame := range frames {
		err.AppendExisting(ErrorAndFrame{Frame: frame})
	}
}

func (err *ErrorsAndFrames) Append(child error, childFrame Frame) {
	err.AppendExisting(ErrorAndFrame{Err: child, Frame: childFrame})
}

func (err *ErrorsAndFrames) AppendExisting(children ...ErrorAndFrame) {
	*err = append(*err, children...)
}

func (err *ErrorsAndFrames) GetErrorsAndFrames() []ErrorAndFrame {
	output := make([]ErrorAndFrame, len(*err))
	copy(output, *err)
	return output
}
