package errors

var (
	Err501NotImplemented      = New("not implemented")
	Err405MethodNotAllowed    = New("method not allowed")
	Err499ClientClosedRequest = New("client closed request")

	// TODO remove
	errStopIteration = New("stop iteration")
)

func Is499ClientClosedRequest(err error) bool {
	return Is(err, Err499ClientClosedRequest)
}

// TODO remove all below

func MakeErrStopIteration() error {
	return errStopIteration
}

func IsStopIteration(err error) bool {
	return Is(err, errStopIteration)
}
