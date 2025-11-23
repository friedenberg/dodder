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

type TypedError[DISAMB any] interface {
	error
	getType() DISAMB
}

func New(text string) TypedError[string] {
	return &errorString[string]{text}
}

func NewWithType[DISAMB any](text string) TypedError[DISAMB] {
	return &errorString[DISAMB]{text}
}

type errorString[DISAMB any] struct {
	value string
}

var _ TypedError[string] = &errorString[string]{}

func (err *errorString[_]) Error() string {
	return err.value
}

func (err *errorString[DISAMB]) getType() (disamb DISAMB) {
	return disamb
}
