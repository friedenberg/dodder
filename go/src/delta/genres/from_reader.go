package genres

import (
	"bufio"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

func ConfirmTypeFromReader(t Genre, r *bufio.Reader) (err error) {
	var t1 Genre

	if t1, err = FromReader(r); err != nil {
		err = errors.Wrap(err)
		return
	}

	if t1 != t {
		err = errors.BadRequest(
			ErrWrongType{
				ExpectedType: t,
				ActualType:   t,
			},
		)
	}

	return
}

func FromReader(r *bufio.Reader) (t Genre, err error) {
	var line string

	if line, err = r.ReadString('\n'); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = t.Set(line); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
