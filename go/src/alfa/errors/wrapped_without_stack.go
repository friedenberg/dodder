package errors

func WithoutStack(err error) errorWithoutStack {
	if err == nil {
		panic("wrapping empty error")
	}

	unwrapped := Unwrap(err)

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

func (err errorWithoutStack) ShouldHideUnwrap() bool {
	return true
}
