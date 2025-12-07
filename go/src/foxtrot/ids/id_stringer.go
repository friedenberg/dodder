package ids

type IdStringerSansRepo struct {
	Id Id
}

func (stringer *IdStringerSansRepo) String() string {
	switch objectId := stringer.Id.(type) {
	case *ObjectId:
		return objectId.StringSansRepo()

	default:
		return objectId.String()
	}
}
