package object_metadata

import "code.linenisgreat.com/dodder/go/src/charlie/markl"

type (
	Lockfile        interface{}
	LockfileMutable interface{}
)

type lockfile struct {
	tags []markl.Id
	tipe markl.Id
}
