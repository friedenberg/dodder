package remote_http

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"github.com/gorilla/mux"
)

type MethodPath struct {
	Method string
	Path   string
}

type Request struct {
	context errors.Context
	request *http.Request
	MethodPath
	Headers http.Header
	Body    io.ReadCloser
}

func (r Request) Vars() map[string]string {
	return mux.Vars(r.request)
}

type Response struct {
	StatusCode int
	Body       io.ReadCloser
}

func (response *Response) ErrorWithStatus(status int, err error) {
	response.StatusCode = status
	response.Body = io.NopCloser(strings.NewReader(err.Error()))
}

func (response *Response) Error(err error) {
	response.ErrorWithStatus(http.StatusInternalServerError, err)
}

func ReadErrorFromBody(response *http.Response) (err error) {
	var sb strings.Builder

	if _, err = io.Copy(&sb, response.Body); err != nil {
		err = errors.ErrorWithStackf(
			"failed to read error string from response (%d) body: %q",
			response.StatusCode,
			err,
		)

		return
	}

	err = errors.BadRequestf(
		"remote responded to request (%q) with error (%d):\n\n%s",
		fmt.Sprintf("%s %s", response.Request.Method, response.Request.URL),
		response.StatusCode,
		&sb,
	)

	return
}

func ReadErrorFromBodyOnGreaterOrEqual(
	response *http.Response,
	status int,
) (err error) {
	if response.StatusCode < status {
		return
	}

	err = ReadErrorFromBody(response)

	return
}

func ReadErrorFromBodyOnNot(
	response *http.Response,
	statuses ...int,
) (err error) {
	if slices.Contains(statuses, response.StatusCode) {
		return
	}

	err = ReadErrorFromBody(response)

	return
}
