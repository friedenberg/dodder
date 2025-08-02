package errors

func WithoutStack(err error) error {
	if err == nil {
		panic("wrapping empty error")
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
