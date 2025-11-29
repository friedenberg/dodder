package ids

import (
	"bytes"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
)

func init() {
	register(config{})
}

var configBytes = []byte("konfig")

func ErrOnConfigBytes(b []byte) (err error) {
	if bytes.Equal(b, configBytes) {
		return errors.ErrorWithStackf("cannot be %q", "konfig")
	}

	return nil
}

func ErrOnConfig(v string) (err error) {
	if v == "konfig" {
		return errors.ErrorWithStackf("cannot be %q", "konfig")
	}

	return nil
}

// TODO turn into singleton
type config struct{}

var Config config

func (a config) IsEmpty() bool {
	return false
}

func (a config) GetGenre() interfaces.Genre {
	return genres.Config
}

func (a config) EqualsAny(b any) bool {
	return values.Equals(a, b)
}

func (a config) Equals(b config) bool {
	return true
}

func (a *config) Reset() {
	return
}

func (a *config) ResetWith(_ config) {
	return
}

func (i config) GetObjectIdString() string {
	return i.String()
}

func (i config) String() string {
	return "konfig"
}

func (k config) Parts() [3]string {
	return [3]string{"", "", "konfig"}
}

func (i config) Set(v string) (err error) {
	v = strings.TrimSpace(v)
	v = strings.ToLower(v)

	if v != "konfig" {
		err = errors.Errorf("not konfig")
		return err
	}

	return err
}

func (t config) MarshalText() (text []byte, err error) {
	text = []byte(t.String())
	return text, err
}

func (t *config) UnmarshalText(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (t config) MarshalBinary() (text []byte, err error) {
	text = []byte(t.String())
	return text, err
}

func (t *config) UnmarshalBinary(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
