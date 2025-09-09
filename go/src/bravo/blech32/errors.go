package blech32

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

var (
	ErrEmptyHRP         = errors.New("empty HRP")
	ErrSeparatorMissing = errors.New(
		fmt.Sprintf("separator (%q) missing", string(separator)),
	)
)

type errInvalidCharacterInData struct {
	pos  int
	char rune
}

func (err errInvalidCharacterInData) Error() string {
	return fmt.Sprintf(
		"invalid character %q found at position %d. expected one of %q",
		string([]rune{err.char}),
		err.pos,
		charsetString,
	)
}
