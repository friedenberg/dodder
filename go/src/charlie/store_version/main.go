package store_version

import (
	"strconv"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
)

var (
	VNull = Version(values.Int(0))
	V1    = Version(values.Int(1))
	V3    = Version(values.Int(3))
	V4    = Version(values.Int(4))
	V6    = Version(values.Int(6))
	V7    = Version(values.Int(7))
	V8    = Version(values.Int(8))
	V9    = Version(values.Int(9))

	// TODO drop support for versions above
	// TODO use golang generation for versions
	V10 = Version(values.Int(10))
	V11 = Version(values.Int(11))
	V12 = Version(values.Int(12))
	V13 = Version(values.Int(13))
	V14 = Version(values.Int(14))

	VCurrent = V12
	VNext    = V13
)

// TODO replace with Int
type Version values.Int

type Getter interface {
	GetStoreVersion() Version
}

func Equals(
	a interfaces.StoreVersion,
	others ...interfaces.StoreVersion,
) bool {
	for _, other := range others {
		if a.GetInt() == other.GetInt() {
			return true
		}
	}

	return false
}

func Less(a, b interfaces.StoreVersion) bool {
	return a.GetInt() < b.GetInt()
}

func LessOrEqual(a, b interfaces.StoreVersion) bool {
	return a.GetInt() <= b.GetInt()
}

func Greater(a, b interfaces.StoreVersion) bool {
	return a.GetInt() > b.GetInt()
}

func GreaterOrEqual(a, b interfaces.StoreVersion) bool {
	return a.GetInt() >= b.GetInt()
}

func (version Version) Less(b interfaces.StoreVersion) bool {
	return Less(version, b)
}

func (version Version) LessOrEqual(b interfaces.StoreVersion) bool {
	return LessOrEqual(version, b)
}

func (version Version) String() string {
	return values.Int(version).String()
}

func (version Version) GetInt() int {
	return values.Int(version).Int()
}

func (version *Version) Set(p string) (err error) {
	var i uint64

	if i, err = strconv.ParseUint(p, 10, 16); err != nil {
		err = errors.Wrap(err)
		return err
	}

	*version = Version(i)

	if VCurrent.Less(version) {
		err = errors.Wrap(ErrFutureStoreVersion{StoreVersion: version})
		return err
	}

	return err
}

func IsVersionLessOrEqualToV11(other interfaces.StoreVersion) bool {
	return LessOrEqual(other, V10)
}

func IsCurrentVersionLessOrEqualToV10() bool {
	return LessOrEqual(VCurrent, V10)
}
