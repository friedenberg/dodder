package values

import (
	"net/url"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type Uri struct {
	url url.URL
}

func (uri *Uri) GetUri() url.URL {
	return uri.url
}

func (uri *Uri) GetUrl() url.URL {
	return uri.url
}

func (uri *Uri) Set(v string) (err error) {
	var u1 *url.URL

	if u1, err = url.Parse(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	uri.url = *u1

	return
}

func (uri *Uri) String() string {
	return uri.url.String()
}

func (uri Uri) MarshalText() (text []byte, err error) {
	text = []byte(uri.String())
	return
}

func (uri *Uri) UnmarshalText(text []byte) (err error) {
	if err = uri.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (uri Uri) MarshalBinary() (text []byte, err error) {
	text = []byte(uri.String())
	return
}

func (uri *Uri) UnmarshalBinary(text []byte) (err error) {
	if err = uri.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
