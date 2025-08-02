package errors

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/http_statuses"
)

func NewHTTPError(statusCode http_statuses.Code) HTTP {
	return HTTP{StatusCode: statusCode}
}

var (
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

func (err HTTP) Errorf(format string, args ...any) HTTP {
	return err.Wrap(fmt.Errorf(format, args...))
}

func (err HTTP) ErrorUnwrappedf(format string, args ...any) HTTP {
	err = err.Wrap(fmt.Errorf(format, args...))
	err.hideUnwrap = true
	return err
}

func (err HTTP) ShouldHideUnwrap() bool {
	return err.hideUnwrap
}
