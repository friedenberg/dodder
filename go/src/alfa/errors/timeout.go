package errors

import (
	"fmt"
	"time"
)

type ErrorOrTimeout struct {
	Timeout time.Duration
	Err     error
}

// func (err ErrorOrTimeout) getNext() (error, funcGetNext) {
// 	var next funcGetNext

// 	switch err := err.Err.(type) {
// 	case UnwrapOne:
// 	case UnwrapMany:
// 	}

// 	return err,
// }

func (err ErrorOrTimeout) Error() string {
	chString := make(chan string)
	defer close(chString)

	go func() {
		chString <- err.Err.Error()
	}()

	select {
	case result := <-chString:
		return result

	case <-time.After(err.Timeout):
		panic(fmt.Sprintf("calling error.Error() timed out for %T", err.Err))
	}
}

func (err ErrorOrTimeout) Is(target error) bool {
	errDanger, ok := err.Err.(ErrorsIs)

	if !ok {
		return false
	}

	chBool := make(chan bool)
	defer close(chBool)

	go func() {
		chBool <- errDanger.Is(target)
	}()

	select {
	case result := <-chBool:
		return result

	case <-time.After(err.Timeout):
		panic(fmt.Sprintf("calling errors.Is() timed out for %T", err.Err))
	}
}

func (err ErrorOrTimeout) Unwrap() []error {
	chErrors := make(chan []error)
	defer close(chErrors)

	go func() {
		switch err := err.Err.(type) {
		case UnwrapOne:
			result := err.Unwrap()
			chErrors <- []error{result}

		case UnwrapMany:
			chErrors <- err.Unwrap()

		default:
			chErrors <- nil
		}
	}()

	select {
	case result := <-chErrors:
		return result

	case <-time.After(err.Timeout):
		panic(fmt.Sprintf("calling error.Unwrap() timed out for %T", err.Err))
	}
}
