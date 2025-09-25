package ids

import (
	"regexp"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

const RepoIdRegexString = `^(/)?[-a-z0-9_]+$`

var RepoIdRegex *regexp.Regexp

func init() {
	RepoIdRegex = regexp.MustCompile(RepoIdRegexString)
	register(RepoId{})
}

func MustRepoId(v string) (e *RepoId) {
	e = &RepoId{}

	if err := e.Set(v); err != nil {
		errors.PanicIfError(err)
	}

	return e
}

func MakeRepoId(v string) (e *RepoId, err error) {
	e = &RepoId{}

	if err = e.Set(v); err != nil {
		err = errors.Wrap(err)
		return e, err
	}

	return e, err
}

type RepoId struct {
	domain string
	id     string
}

func (k RepoId) IsEmpty() bool {
	return k.id == ""
}

func (k RepoId) GetRepoId() interfaces.RepoId {
	return k
}

func (k RepoId) EqualsRepoId(kg interfaces.RepoIdGetter) bool {
	return kg.GetRepoId().GetRepoIdString() == k.GetRepoIdString()
}

func (k RepoId) GetRepoIdString() string {
	return k.String()
}

func (e *RepoId) Reset() {
	e.id = ""
}

func (e *RepoId) ResetWith(e1 RepoId) {
	e.id = e1.id
}

func (a RepoId) EqualsAny(b any) bool {
	return values.Equals(a, b)
}

func (a RepoId) Equals(b RepoId) bool {
	return a.id == b.id
}

func (o RepoId) GetGenre() interfaces.Genre {
	return genres.Repo
}

func (i RepoId) GetObjectIdString() string {
	return i.String()
}

func (k RepoId) String() string {
	return k.id
}

func (k RepoId) StringWithSlashPrefix() string {
	return "/" + k.id
}

func (k RepoId) Parts() [3]string {
	return [3]string{"", "/", k.id}
}

func (k RepoId) GetQueryPrefix() string {
	return "/"
}

func (e *RepoId) Set(v string) (err error) {
	v = strings.TrimPrefix(v, "/")
	v = strings.ToLower(strings.TrimSpace(v))

	if err = ErrOnConfig(v); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if v == "" {
		return err
	}

	if !RepoIdRegex.Match([]byte(v)) {
		err = errors.ErrorWithStackf("not a valid Kasten: '%s'", v)
		return err
	}

	e.id = v

	return err
}

func (t RepoId) MarshalText() (text []byte, err error) {
	text = []byte(t.String())
	return text, err
}

func (t *RepoId) UnmarshalText(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (t RepoId) MarshalBinary() (text []byte, err error) {
	text = []byte(t.String())
	return text, err
}

func (t *RepoId) UnmarshalBinary(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
