package ids

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
)

func MakeExternalObjectId(genre genres.Genre, value string) *ExternalObjectId {
	return &ExternalObjectId{
		value: value,
		genre: genre,
	}
}

type ExternalObjectId struct {
	value string
	genre genres.Genre
}

var _ Id = &ExternalObjectId{}

func (id *ExternalObjectId) GetExternalObjectId() interfaces.ExternalObjectId {
	return id
}

func (id *ExternalObjectId) GetGenre() interfaces.Genre {
	return id.genre
}

func (id *ExternalObjectId) IsEmpty() bool {
	return id.value == ""
}

func (id *ExternalObjectId) String() string {
	return id.value
}

func (id *ExternalObjectId) SetGenre(genre interfaces.Genre) (err error) {
	id.genre = genres.Must(genre)
	return err
}

func (id *ExternalObjectId) SetBlob(v string) (err error) {
	id.genre = genres.Blob

	if err = id.Set(v); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (id *ExternalObjectId) Set(value string) (err error) {
	if value == "/" {
		id.Reset()
		return err
	}

	if len(value) <= 1 {
		err = errors.ErrorWithStackf("external object id must be at least two characters, but got %q", value)
		return err
	}

	id.value = value

	return err
}

func (id *ExternalObjectId) SetWithGenre(
	value string,
	genre interfaces.Genre,
) (err error) {
	if err = id.Set(value); err != nil {
		err = errors.Wrap(err)
		return err
	}

	id.genre = genres.Must(genre)
	return err
}

func (id *ExternalObjectId) Reset() {
	id.genre = genres.None
	id.value = ""
}

func (id *ExternalObjectId) ResetWith(src *ExternalObjectId) {
	id.genre = src.genre
	id.value = src.value
}

func (id *ExternalObjectId) SetObjectIdLike(other Id) (err error) {
	if other.IsEmpty() {
		id.Reset()
		return err
	}

	var value string

	if oid, ok := other.(*ObjectId); ok {
		value = oid.StringSansOp()
	} else {
		value = other.String()
	}

	if err = id.SetWithGenre(
		value,
		genres.Must(other.GetGenre()),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (id *ExternalObjectId) MarshalBinary() (b []byte, err error) {
	if b, err = id.genre.MarshalBinary(); err != nil {
		err = errors.Wrap(err)
		return b, err
	}

	b = append(b, []byte(id.value)...)

	return b, err
}

func (id *ExternalObjectId) UnmarshalBinary(b []byte) (err error) {
	if err = id.genre.UnmarshalBinary(b[:1]); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if len(b) > 1 {
		id.value = string(b[1:])
	}

	return err
}

func (id *ExternalObjectId) ToType() TypeStruct {
	panic(errors.Err405MethodNotAllowed)
}

func (id *ExternalObjectId) ToSeq() doddish.Seq {
	panic(errors.Err501NotImplemented)
}
