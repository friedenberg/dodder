package genres

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
)

var ErrNoAbbreviation = errors.New("no abbreviation")

func MakeErrUnsupportedGenre(g interfaces.GenreGetter) error {
	return errors.WrapSkip(1, errUnsupportedGenre{Genre: g.GetGenre()})
}

func IsErrUnsupportedGenre(err error) bool {
	return errors.Is(err, errUnsupportedGenre{Genre: None})
}

type errUnsupportedGenre struct {
	interfaces.Genre
}

func (e errUnsupportedGenre) Is(target error) (ok bool) {
	_, ok = target.(errUnsupportedGenre)
	return ok
}

func (e errUnsupportedGenre) Error() string {
	return fmt.Sprintf("unsupported genre: %q", e.Genre)
}

func MakeErrUnrecognizedGenre(v string) errUnrecognizedGenre {
	return errUnrecognizedGenre(v)
}

func IsErrUnrecognizedGenre(err error) bool {
	return errors.Is(err, errUnrecognizedGenre(""))
}

type errUnrecognizedGenre string

func (e errUnrecognizedGenre) Is(target error) (ok bool) {
	_, ok = target.(errUnrecognizedGenre)
	return ok
}

func (e errUnrecognizedGenre) Error() string {
	return fmt.Sprintf(
		"unknown genre: %q. Available genres: %q",
		string(e),
		quiter.Strings(collections_slice.Slice[Genre](All())),
	)
}

type ErrWrongType struct {
	ExpectedType, ActualType Genre
}

func (e ErrWrongType) Is(target error) (ok bool) {
	_, ok = target.(ErrWrongType)
	return ok
}

func (e ErrWrongType) Error() string {
	return fmt.Sprintf(
		"expected zk_types %s but got %s",
		e.ExpectedType,
		e.ActualType,
	)
}

type ErrEmptyObjectId struct{}

func (e ErrEmptyObjectId) Is(target error) (ok bool) {
	_, ok = target.(ErrEmptyObjectId)
	return ok
}

func (e ErrEmptyObjectId) Error() string {
	return fmt.Sprintf("empty object id")
}
