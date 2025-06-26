package errors

var (
	ErrNotImplemented = New("not implemented")
	ErrNotSupported   = New("not supported")
)

// TODO remove all below

var (
	ErrFalse         = New("false")
	ErrTrue          = New("true")
	ErrStopIteration = New("stop iteration")
)

func MakeErrStopIteration() error {
	return ErrStopIteration
}

func IsStopIteration(err error) bool {
	if Is(err, ErrStopIteration) {
		return true
	}

	return false
}

func IsErrFalse(err error) bool {
	if Is(err, ErrFalse) {
		return true
	}

	return false
}

func IsErrTrue(err error) bool {
	if Is(err, ErrTrue) {
		return true
	}

	return false
}
