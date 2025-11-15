package object_metadata

import "code.linenisgreat.com/dodder/go/src/charlie/markl"

type (
	Lockfile interface {
		GetTypeLock() Lock
	}

	LockfileMutable interface{}

	Lock struct {
		Key string
		Id  markl.Id
	}
)

type lockfile struct {
	tipe Lock
	tags []Lock
}

var _ Lockfile = lockfile{}

func (lockfile lockfile) GetTypeLock() Lock {
	return lockfile.tipe
}
