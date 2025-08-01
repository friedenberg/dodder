package errors

// TODO redesign all the below
var errStopIteration = New("stop iteration")

func MakeErrStopIteration() error {
	return errStopIteration
}

func IsStopIteration(err error) bool {
	ok := Is(err, errStopIteration)
	return ok
}
