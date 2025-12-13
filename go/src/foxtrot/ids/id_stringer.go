package ids

type StringerSansRepo struct {
	Id Id
}

func (stringer *StringerSansRepo) String() string {
	switch objectId := stringer.Id.(type) {
	case *ObjectId:
		return objectId.StringSansRepo()

	default:
		return objectId.String()
	}
}
