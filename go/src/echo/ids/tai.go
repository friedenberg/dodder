package ids

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/charlie/delim_io"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/delta/thyme"
	chai "github.com/brandondube/tai"
)

type tai = chai.TAI

func init() {
	register(Tai{})
	collections_value.RegisterGobValue[Tai](nil)
}

type Tai struct {
	wasSet bool
	tai
}

func NowTai() Tai {
	return Tai{
		wasSet: true,
		tai:    chai.Now(),
	}
}

func TaiFromTime(t1 thyme.Time) (t2 Tai) {
	t2 = TaiFromTimeWithIndex(t1, 0)
	return
}

func TaiFromTime1(t1 time.Time) (t2 Tai) {
	t2.wasSet = true
	t2.tai = chai.FromTime(t1)

	return
}

func TaiFromTimeWithIndex(t1 thyme.Time, n int) (t2 Tai) {
	t2.wasSet = true
	t2.tai = chai.FromTime(t1.GetTime())
	t2.Asec += int64(n * chai.Attosecond)

	return
}

func (tai Tai) AsTime() (t1 thyme.Time) {
	t1 = thyme.Tyme(tai.tai.AsTime().Local())
	return
}

func (tai Tai) Before(b Tai) bool {
	return tai.tai.Before(b.tai)
}

func (tai Tai) After(b Tai) bool {
	return tai.tai.After(b.tai)
}

func (tai Tai) GetGenre() interfaces.Genre {
	return genres.InventoryList
}

func (tai Tai) Parts() [3]string {
	a := strings.TrimRight(fmt.Sprintf("%018d", tai.Asec), "0")

	if a == "" {
		a = "0"
	}

	return [3]string{strconv.FormatInt(tai.Sec, 10), ".", a}
}

func (tai Tai) GetObjectIdString() string {
	return tai.String()
}

func (tai Tai) String() (v string) {
	a := strings.TrimRight(fmt.Sprintf("%018d", tai.Asec), "0")

	if a == "" {
		a = "0"
	}

	v = fmt.Sprintf("%s.%s", strconv.FormatInt(tai.Sec, 10), a)

	// if v == "0.0" {
	// 	panic("empty tai")
	// }

	return
}

func (tai Tai) StringDefaultFormat() string {
	f := string_format_writer.StringFormatDateTime + ".000000000"
	return tai.Format(f)
}

func (tai Tai) Format(v string) string {
	return tai.AsTime().Format(v)
}

func (tai *Tai) SetFromRFC3339(v string) (err error) {
	tai.wasSet = true

	var t1 time.Time

	if t1, err = thyme.Parse(thyme.RFC3339, v); err != nil {
		err = errors.Wrap(err)
		return
	}

	*tai = TaiFromTime1(t1)

	return
}

func (tai *Tai) Set(value string) (err error) {
	tai.wasSet = true

	reader, repool := pool.GetStringReader(value)
	defer repool()

	delimiterReader := delim_io.Make('.', reader)
	defer delim_io.PutReader(delimiterReader)

	idx := 0
	var val string

	var isEOF bool

	for !isEOF {
		val, err = delimiterReader.ReadOneString()

		if err == io.EOF {
			isEOF = true
			err = nil
		} else if err != nil {
			err = errors.Wrap(err)
			return
		}

		switch idx {
		case 0:
			val = strings.TrimSpace(val)

			if len(val) == 0 {
				break
			}

			if tai.Sec, err = strconv.ParseInt(val, 10, 64); err != nil {
				err = errors.Wrapf(err, "failed to parse Sec time: %s", value)
				return
			}

		case 1:
			val = strings.TrimSpace(val)
			val = strings.TrimRight(val, "0")

			if val == "" {
				break
			}

			var pre int64

			if pre, err = strconv.ParseInt(val, 10, 64); err != nil {
				err = errors.Wrapf(err, "failed to parse Asec time: %s", val)
				return
			}

			tai.Asec = pre * int64(math.Pow10(18-len(val)))

		default:
			err = errors.ErrorWithStackf(
				"expected no more elements but got %s",
				val,
			)
			return
		}

		idx++
	}

	return
}

func (tai Tai) GetBlobId() interfaces.MarklId {
	return markl.HashTypeSha256.FromStringContent(tai.String())
}

func (tai Tai) IsZero() (ok bool) {
	ok = (tai.Sec == 0 && tai.Asec == 0) || !tai.wasSet
	return
}

func (tai Tai) IsEmpty() (ok bool) {
	ok = tai.IsZero()
	return
}

func (tai Tai) GetTai() Tai {
	return tai
}

func (tai *Tai) Reset() {
	tai.Sec = 0
	tai.Asec = 0
	tai.wasSet = false
}

func (tai *Tai) ResetWith(b Tai) {
	tai.Sec = b.Sec
	tai.Asec = b.Asec
	tai.wasSet = b.wasSet
}

func (tai Tai) WriteTo(w io.Writer) (n int64, err error) {
	b := make([]byte, binary.MaxVarintLen64*2)
	binary.PutVarint(b[:binary.MaxVarintLen64], tai.Sec)
	binary.PutVarint(b[binary.MaxVarintLen64:], tai.Asec)
	var n1 int
	n1, err = ohio.WriteAllOrDieTrying(w, b)
	n += int64(n1)
	return
}

func (tai *Tai) ReadFrom(r io.Reader) (n int64, err error) {
	b := make([]byte, binary.MaxVarintLen64*2)

	var n1 int
	n1, err = ohio.ReadAllOrDieTrying(r, b)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	tai.wasSet = true
	tai.Sec, _ = binary.Varint(b[:binary.MaxVarintLen64])
	tai.Asec, _ = binary.Varint(b[binary.MaxVarintLen64:])

	return
}

func (tai Tai) MarshalText() (text []byte, err error) {
	ui.Err().Printf(tai.String())
	text = []byte(tai.String())

	return
}

func (tai *Tai) UnmarshalText(text []byte) (err error) {
	if err = tai.Set(string(text)); err != nil {
		return
	}

	return
}

func (tai Tai) MarshalBinary() (text []byte, err error) {
	text = []byte(tai.String())

	return
}

func (tai *Tai) UnmarshalBinary(text []byte) (err error) {
	if err = tai.Set(string(text)); err != nil {
		return
	}

	return
}

func (tai Tai) EqualsAny(b any) bool {
	return values.Equals(tai, b)
}

func (tai Tai) Equals(t1 Tai) bool {
	if !tai.Eq(t1.tai) {
		return false
	}

	return true
}

func (tai Tai) Less(t1 Tai) bool {
	return tai.Before(t1)
}

func (tai Tai) SortCompare(t1 Tai) quiter.SortCompare {
	if tai.Equals(t1) {
		return quiter.SortCompareEqual
	} else if tai.Before(t1) {
		return quiter.SortCompareLess
	} else {
		return quiter.SortCompareGreater
	}
}

func MakeTaiRFC3339Value(t Tai) *TaiRFC3339Value {
	t1 := TaiRFC3339Value(t)
	return &t1
}

type TaiRFC3339Value Tai

func (t *TaiRFC3339Value) Set(v string) (err error) {
	t.wasSet = true

	var t1 time.Time

	if t1, err = thyme.Parse(thyme.RFC3339, v); err != nil {
		err = errors.Wrap(err)
		return
	}

	*t = TaiRFC3339Value(TaiFromTime1(t1))

	return
}

func (t *TaiRFC3339Value) String() string {
	// TODO figure out why the pointer needs to be converted to Tai to execute
	// correctly
	return Tai(*t).Format(thyme.RFC3339)
}
