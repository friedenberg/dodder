package errors

import "testing"

func TestBadRequestf(t *testing.T) {
	badRequest := BadRequestf("testing")
	badRequest = Wrap(badRequest)

	if !Is400BadRequest(badRequest) {
		t.Errorf("expected bad request")
	}
}

func TestBadRequest(t *testing.T) {
	badRequest := BadRequest(New("testing"))
	badRequest = Wrap(badRequest)

	if !Is400BadRequest(badRequest) {
		t.Errorf("expected bad request")
	}
}
