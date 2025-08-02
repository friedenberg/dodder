package errors

import "testing"

func TestBadRequestf(t *testing.T) {
	var badRequest error = BadRequestf("testing")
	badRequest = Wrap(badRequest)

	if !IsBadRequest(badRequest) {
		t.Errorf("expected bad request")
	}
}

func TestBadRequest(t *testing.T) {
	var badRequest error = BadRequest(New("testing"))
	badRequest = Wrap(badRequest)

	if !IsBadRequest(badRequest) {
		t.Errorf("expected bad request")
	}
}
