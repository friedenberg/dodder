package remote_http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/mcp"
	"code.linenisgreat.com/dodder/go/src/charlie/error_coders"
	"github.com/gorilla/mux"
)

type MethodPath struct {
	Method string
	Path   string
}

type Request struct {
	ctx     interfaces.Context
	request *http.Request
	MethodPath
	Headers http.Header
	Body    io.ReadCloser
}

func (request Request) Vars() map[string]string {
	return mux.Vars(request.request)
}

type Response struct {
	StatusCode int
	headers    http.Header
	Body       io.ReadCloser
}

func (response *Response) Headers() http.Header {
	if response.headers == nil {
		response.headers = make(http.Header)
	}

	return response.headers
}

func (response *Response) ErrorWithStatus(status int, err error) {
	response.StatusCode = status

	if err != nil {
		var buffer bytes.Buffer

		error_coders.Encoder.EncodeTo(err, &buffer)
		response.Body = io.NopCloser(&buffer)
	}
}

func (response *Response) Error(err error) {
	response.ErrorWithStatus(http.StatusInternalServerError, err)
}

func (response *Response) MCPError(
	status int,
	id any,
	code int,
	message string,
	data any,
) {
	response.StatusCode = status

	mcpResponse := mcp.Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &mcp.Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	responseBytes, _ := json.Marshal(mcpResponse)
	response.Body = io.NopCloser(bytes.NewReader(responseBytes))
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

	body := sb.String()

	endpointText := fmt.Sprintf(
		"%s %s",
		response.Request.Method,
		response.Request.URL,
	)

	statusText := fmt.Sprintf(
		"%d %s", response.StatusCode,
		http.StatusText(response.StatusCode),
	)

	if body == "" {
		err = errors.BadRequestf(
			"remote responded to request (%q) with error status: %s (error body not provided)",
			endpointText,
			statusText,
		)
	} else {
		err = errors.BadRequestf(
			"remote responded to request (%q) with error (%s):\n\n%s",
			endpointText,
			statusText,
			body,
		)
	}

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
