package stack_frame

import "errors"

func ErrorWithoutStack(err error) errorWithoutStack {
	unwrapped := errors.Unwrap(err)

	if unwrapped != nil {
		err = unwrapped
	}

	return errorWithoutStack{underlying: err}
}

type errorWithoutStack struct {
	underlying error
}

func (err errorWithoutStack) Error() string {
	return err.underlying.Error()
}

func (err errorWithoutStack) Unwrap() error {
	return err.underlying
}

func (err errorWithoutStack) ShouldShowStackTrace() bool {
	return false
}
