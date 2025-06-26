package errors

import (
	"strings"

	"golang.org/x/xerrors"
)

// TODO combine all of the below errors into a more useful grouping

type StackTracer interface {
	error
	ShouldShowStackTrace() bool
}

type ErrBadRequest interface {
	IsBadRequest()
}

func BadRequest(err error) *errBadRequestWrap {
	return &errBadRequestWrap{err}
}

func BadRequestf(fmt string, args ...any) *errBadRequestWrap {
	return &errBadRequestWrap{xerrors.Errorf(fmt, args...)}
}

func BadRequestPrefix(preamble string, err error) *errBadRequestPreamble {
	return &errBadRequestPreamble{
		preamble: preamble,
		error:    err,
	}
}

func IsBadRequest(err error) bool {
	return Is(err, errBadRequestWrap{}) || Is(err, errBadRequestPreamble{})
}

type errBadRequestPreamble struct {
	preamble string
	error
}

func (err errBadRequestPreamble) IsBadRequest() {}

func (e errBadRequestPreamble) ShouldShowStackTrace() bool {
	return false
}

func (e errBadRequestPreamble) Is(target error) bool {
	_, ok := target.(ErrBadRequest)
	return ok
}

func (err errBadRequestPreamble) Error() string {
	var sb strings.Builder
	sb.WriteString(err.preamble)
	sb.WriteString(": \n\n")

	if stackWrapError, ok := err.error.(*stackWrapError); ok {
		stackWrapError.writeErrorNoStack(&sb)
	} else {
		sb.WriteString(err.error.Error())
	}

	return sb.String()
}

type errBadRequestWrap struct {
	error
}

func (err errBadRequestWrap) IsBadRequest() {}

func (e errBadRequestWrap) ShouldShowStackTrace() bool {
	return false
}

func (e errBadRequestWrap) Is(target error) bool {
	_, ok := target.(ErrBadRequest)
	return ok
}

func (e errBadRequestWrap) Error() string {
	return e.error.Error()
}

// TODO refactor NewNormal into something that combines helpful and stack trace
func NewNormal(v string) errNormal {
	return errNormal{string: v}
}

type errNormal struct {
	string
}

func (e errNormal) ShouldShowStackTrace() bool {
	return false
}

func (e errNormal) Is(target error) bool {
	_, ok := target.(errNormal)
	return ok
}

func (e errNormal) Error() string {
	return e.string
}
