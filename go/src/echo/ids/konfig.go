package ids

import (
	"bytes"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

func init() {
	register(Config{})
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

type Config struct{}

func (a Config) IsEmpty() bool {
	return false
}

func (a Config) GetGenre() interfaces.Genre {
	return genres.Config
}

func (a Config) EqualsAny(b any) bool {
	return values.Equals(a, b)
}

func (a Config) Equals(b Config) bool {
	return true
}

func (a *Config) Reset() {
	return
}

func (a *Config) ResetWith(_ Config) {
	return
}

func (i Config) GetObjectIdString() string {
	return i.String()
}

func (i Config) String() string {
	return "konfig"
}

func (k Config) Parts() [3]string {
	return [3]string{"", "", "konfig"}
}

func (i Config) Set(v string) (err error) {
	v = strings.TrimSpace(v)
	v = strings.ToLower(v)

	if v != "konfig" {
		err = errors.Errorf("not konfig")
		return err
	}

	return err
}

func (t Config) MarshalText() (text []byte, err error) {
	text = []byte(t.String())
	return text, err
}

func (t *Config) UnmarshalText(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (t Config) MarshalBinary() (text []byte, err error) {
	text = []byte(t.String())
	return text, err
}

func (t *Config) UnmarshalBinary(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
