package values

import (
	"net/url"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type Uri struct {
	url url.URL
}

func (u *Uri) GetUrl() url.URL {
	return u.url
}

func (u *Uri) Set(v string) (err error) {
	var u1 *url.URL

	if u1, err = url.Parse(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	u.url = *u1

	return
}

func (u *Uri) String() string {
	return u.url.String()
}

func (u Uri) MarshalText() (text []byte, err error) {
	text = []byte(u.String())
	return
}

func (u *Uri) UnmarshalText(text []byte) (err error) {
	if err = u.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (u Uri) MarshalBinary() (text []byte, err error) {
	text = []byte(u.String())
	return
}

func (u *Uri) UnmarshalBinary(text []byte) (err error) {
	if err = u.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
