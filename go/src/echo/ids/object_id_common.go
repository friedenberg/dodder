package ids

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

type ObjectIdGetter interface {
	GetObjectId() *ObjectId
}

func GetObjectIdPool() interfaces.Pool[ObjectId, *ObjectId] {
	return getObjectIdPool2()
}

type ObjectId = objectId2

type IdParts struct {
	Middle              byte
	RepoId, Left, Right *catgut.String
}

var ErrFDNotId = errors.New("not a id file")

func MustObjectId(kp interfaces.ObjectId) (k *ObjectId) {
	k = &ObjectId{}
	err := k.SetWithIdLike(kp)
	errors.PanicIfError(err)
	return
}

type ObjectIdStringerSansRepo struct {
	ObjectIdLike
}

func (oid *ObjectIdStringerSansRepo) String() string {
	switch oid := oid.ObjectIdLike.(type) {
	case *ObjectId:
		return oid.StringSansRepo()

	default:
		return oid.String()
	}
}

type ObjectIdStringerWithRepo ObjectId

func (oid *ObjectIdStringerWithRepo) String() string {
	var sb strings.Builder

	if oid.repoId.Len() > 0 {
		sb.WriteRune('/')
		oid.repoId.WriteTo(&sb)
		sb.WriteRune('/')
	}

	switch oid.g {
	case genres.Zettel:
		sb.Write(oid.left.Bytes())

		if oid.middle != '\x00' {
			sb.WriteByte(oid.middle)
		}

		sb.Write(oid.right.Bytes())

	case genres.Type:
		sb.Write(oid.right.Bytes())

	default:
		if oid.left.Len() > 0 {
			sb.Write(oid.left.Bytes())
		}

		if oid.middle != '\x00' {
			sb.WriteByte(oid.middle)
		}

		if oid.right.Len() > 0 {
			sb.Write(oid.right.Bytes())
		}
	}

	return sb.String()
}
