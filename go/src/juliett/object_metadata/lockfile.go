package object_metadata

import "code.linenisgreat.com/dodder/go/src/foxtrot/markl"

type (
	Lockfile interface {
		GetType() markl.Id
	}

	LockfileMutable interface {
		Lockfile
	}

	Lock struct {
		Key string
		Id  markl.Id
	}
)

type lockfile struct {
	Type markl.Id
	tags []Lock
}

var (
	_ Lockfile        = lockfile{}
	_ LockfileMutable = &lockfile{}
)

func (lockfile lockfile) GetType() markl.Id {
	return lockfile.Type
}
