package errors

import (
	"fmt"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/http_statuses"
)

func NewHTTPError(statusCode http_statuses.Code) HTTP {
	return HTTP{StatusCode: statusCode}
}

func BadRequest(err error) error {
	fmt.Fprintf(os.Stderr, "%#v\n", err)
	if Is400BadRequest(err) {
		return err
	}

	return WithoutStack(Err400BadRequest.WrapHidden(err))
}

func BadRequestf(format string, args ...any) error {
	return WithoutStack(Err400BadRequest.ErrorHiddenf(format, args...))
}

var (
	Err400BadRequest       = NewHTTPError(http_statuses.Code400BadRequest)
	Err405MethodNotAllowed = NewHTTPError(
		http_statuses.Code405MethodNotAllowed,
	)
	Err409Conflict            = NewHTTPError(http_statuses.Code409Conflict)
	Err499ClientClosedRequest = NewHTTPError(
		http_statuses.Code499ClientClosedRequest,
	)
	Err501NotImplemented = NewHTTPError(
		http_statuses.Code501NotImplemented,
	)
)

func Is400BadRequest(err error) bool {
	return IsHTTPError(err, http_statuses.Code400BadRequest)
}

func Is499ClientClosedRequest(err error) bool {
	return IsHTTPError(err, http_statuses.Code499ClientClosedRequest)
}

func IsHTTPError(target error, statusCode http_statuses.Code) bool {
	var errHTTP HTTP

	if !As(target, &errHTTP) {
		return false
	}

	return errHTTP.StatusCode == statusCode
}

type HTTP struct {
	StatusCode http_statuses.Code
	hideUnwrap bool
	underlying error
}

func (err HTTP) Error() string {
	return err.StatusCode.String()
}

func (err HTTP) Is(target error) bool {
	_, ok := target.(HTTP)
	return ok
}

func (err HTTP) Unwrap() error {
	return err.underlying
}

func (err HTTP) Wrap(underlying error) HTTP {
	return HTTP{
		StatusCode: err.StatusCode,
		underlying: underlying,
	}
}

func (err HTTP) WrapHidden(underlying error) HTTP {
	return HTTP{
		StatusCode: err.StatusCode,
		underlying: underlying,
		hideUnwrap: true,
	}
}

func (err HTTP) Errorf(format string, args ...any) HTTP {
	err = err.WrapHidden(fmt.Errorf(format, args...))
	return err
}

// Creates a new error from `format` and `args` a returns a new HTTP error that
// wraps it, but does not expose the HTTP error's message.
//
// The returned error will satisfy the appropriate `IsHTTPError(err, status)`
// call, but when using `error_coders` to print it, but it won't show the HTTP
// error
func (err HTTP) ErrorHiddenf(format string, args ...any) HTTP {
	err = err.WrapHidden(fmt.Errorf(format, args...))
	return err
}

func (err HTTP) ShouldHideUnwrap() bool {
	return err.hideUnwrap
}
