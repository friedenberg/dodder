package ids

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
)

type ObjectIdGetter interface {
	GetObjectId() *ObjectId
}

type ObjectIdStringMarshalerSansRepo struct {
	interfaces.ObjectId
}

func (objectId *ObjectIdStringMarshalerSansRepo) String() string {
	switch objectId := objectId.ObjectId.(type) {
	case *ObjectId:
		return objectId.StringSansRepo()

	default:
		return objectId.String()
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

	switch oid.genre {
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
