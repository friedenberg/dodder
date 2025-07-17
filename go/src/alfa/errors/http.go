package errors

import (
	"fmt"
	"net/http"
)

type HTTP struct {
	StatusCode int
}

func (err HTTP) Error() string {
	text := http.StatusText(err.StatusCode)
	return fmt.Sprintf("%d %s", err.StatusCode, text)
}

func (err HTTP) Is(target error) bool {
	_, ok := target.(HTTP)
	return ok
}
