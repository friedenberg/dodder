package errors

var (
	Err501NotImplemented      = New("not implemented")
	Err405MethodNotAllowed    = New("method not allowed")
	Err499ClientClosedRequest = New("client closed request")
)

func Is499ClientClosedRequest(err error) bool {
	return Is(err, Err499ClientClosedRequest)
}

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
	return Is(err, ErrStopIteration)
}
