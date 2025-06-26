package ids

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

func MakeExternalObjectId(g genres.Genre, value string) *ExternalObjectId {
	return &ExternalObjectId{
		value: value,
		genre: g,
	}
}

type ExternalObjectId struct {
	value string
	genre genres.Genre
}

func (eoid *ExternalObjectId) GetExternalObjectId() ExternalObjectIdLike {
	return eoid
}

func (eoid *ExternalObjectId) GetGenre() interfaces.Genre {
	return eoid.genre
}

func (eoid *ExternalObjectId) IsEmpty() bool {
	return eoid.value == ""
}

func (eoid *ExternalObjectId) String() string {
	return eoid.value
}

func (eoid *ExternalObjectId) SetGenre(genre interfaces.Genre) (err error) {
	eoid.genre = genres.Must(genre)
	return
}

func (eoid *ExternalObjectId) SetBlob(v string) (err error) {
	eoid.genre = genres.Blob

	if err = eoid.Set(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (eoid *ExternalObjectId) Set(value string) (err error) {
	if value == "/" {
		eoid.Reset()
		return
	}

	if len(value) <= 1 {
		err = errors.ErrorWithStackf("external object id must be at least two characters, but got %q", value)
		return
	}

	eoid.value = value

	return
}

func (eoid *ExternalObjectId) SetWithGenre(
	value string,
	genre interfaces.Genre,
) (err error) {
	if err = eoid.Set(value); err != nil {
		err = errors.Wrap(err)
		return
	}

	eoid.genre = genres.Must(genre)
	return
}

func (eoid *ExternalObjectId) Reset() {
	eoid.genre = genres.None
	eoid.value = ""
}

func (dst *ExternalObjectId) ResetWith(src *ExternalObjectId) {
	dst.genre = src.genre
	dst.value = src.value
}

func (dst *ExternalObjectId) SetObjectIdLike(src ObjectIdLike) (err error) {
	if src.IsEmpty() {
		dst.Reset()
		return
	}

	var value string

	if oid, ok := src.(*ObjectId); ok {
		value = oid.StringSansOp()
	} else {
		value = src.String()
	}

	if err = dst.SetWithGenre(
		value,
		genres.Must(src.GetGenre()),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (eoid *ExternalObjectId) MarshalBinary() (b []byte, err error) {
	if b, err = eoid.genre.MarshalBinary(); err != nil {
		err = errors.Wrap(err)
		return
	}

	b = append(b, []byte(eoid.value)...)

	return
}

func (eoid *ExternalObjectId) UnmarshalBinary(b []byte) (err error) {
	if err = eoid.genre.UnmarshalBinary(b[:1]); err != nil {
		err = errors.Wrap(err)
		return
	}

	if len(b) > 1 {
		eoid.value = string(b[1:])
	}

	return
}
