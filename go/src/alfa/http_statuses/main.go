package http_statuses

import (
	"fmt"
	"net/http"
)

type Code int

func (code Code) String() string {
	text := http.StatusText(int(code))
	return fmt.Sprintf("%d %s", code, text)
}

const (
	Code405MethodNotAllowed    = Code(http.StatusMethodNotAllowed)
	Code409Conflict            = Code(http.StatusConflict)
	Code499ClientClosedRequest = Code(499)
	Code501NotImplemented      = Code(http.StatusNotImplemented)
)
