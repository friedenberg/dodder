package sku

type transactedLessorTaiOnly struct{}

func (transactedLessorTaiOnly) Less(a, b *Transacted) bool {
	return a.GetTai().Less(b.GetTai())
}

func (transactedLessorTaiOnly) LessPtr(a, b *Transacted) bool {
	return a.GetTai().Less(b.GetTai())
}

type transactedLessorStable struct{}

func (transactedLessorStable) Less(a, b *Transacted) bool {
	if result := a.GetTai().SortCompare(b.GetTai()); !result.Equal() {
		return result.Less()
	}

	return a.GetObjectId().String() < b.GetObjectId().String()
}

func (transactedLessorStable) LessPtr(a, b *Transacted) bool {
	return a.GetTai().Less(b.GetTai())
}

type transactedEqualer struct{}

func (transactedEqualer) Equals(a, b *Transacted) bool {
	return a.Equals(b)
}
