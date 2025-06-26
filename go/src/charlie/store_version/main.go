package store_version

import (
	"strconv"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
)

const currentVersion = 9

var (
	V1  = Version(values.Int(1))
	V3  = Version(values.Int(3))
	V4  = Version(values.Int(4))
	V6  = Version(values.Int(6))
	V7  = Version(values.Int(7))
	V8  = Version(values.Int(8))
	V9  = Version(values.Int(9))
	V10 = Version(values.Int(10))

	VCurrent = V10
	VNext    = V10
)

type Version values.Int

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

func (a Version) Less(b interfaces.StoreVersion) bool {
	return Less(a, b)
}

func (a Version) LessOrEqual(b interfaces.StoreVersion) bool {
	return LessOrEqual(a, b)
}

func (a Version) String() string {
	return values.Int(a).String()
}

func (a Version) GetInt() int {
	return values.Int(a).Int()
}

func (v *Version) Set(p string) (err error) {
	var i uint64

	if i, err = strconv.ParseUint(p, 10, 16); err != nil {
		err = errors.Wrap(err)
		return
	}

	*v = Version(i)

	if VCurrent.Less(v) {
		err = errors.Wrap(ErrFutureStoreVersion{StoreVersion: v})
		return
	}

	return
}

// func (v *StoreVersion) ReadFromFile(
// 	p string,
// ) (err error) {
// 	if err = v.ReadFromFileOrVersion(p, StoreVersionCurrent); err != nil {
// 		err = errors.Wrap(err)
// 		return
// 	}

// 	return
// }

// func (v *StoreVersion) ReadFromFileOrVersion(
// 	p string,
// 	alternative StoreVersion,
// ) (err error) {
// 	var b []byte

// 	var f *os.File

// 	if f, err = files.Open(p); err != nil {
// 		if errors.IsNotExist(err) {
// 			*v = alternative
// 			err = nil
// 		} else {
// 			err = errors.Wrap(err)
// 		}

// 		return
// 	}

// 	if b, err = io.ReadAll(f); err != nil {
// 		err = errors.Wrap(err)
// 		return
// 	}

// 	if err = v.Set(string(b)); err != nil {
// 		err = errors.Wrap(err)
// 		return
// 	}

// 	return
// }
