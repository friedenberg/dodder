package remote_http

import (
	"bufio"
	"net/http"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type roundTripperBufio struct {
	*bufio.Writer
	*bufio.Reader
}

func (roundTripper *roundTripperBufio) RoundTrip(
	request *http.Request,
) (response *http.Response, err error) {
	if err = request.Write(roundTripper.Writer); err != nil {
		err = errors.ErrorWithStackf("failed to write to socket: %w", err)
		return
	}

	if err = roundTripper.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if response, err = http.ReadResponse(
		roundTripper.Reader,
		request,
	); err != nil {
		err = errors.ErrorWithStackf("failed to read response: %w", err)
		return
	}

	return
}
