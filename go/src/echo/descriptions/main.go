package descriptions

import (
	"io"
	"strings"
	"unicode/utf8"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/doddish"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
)

// TODO-P1 move to catgut.String
type Description struct {
	wasSet bool
	value  string
}

func Make(v string) Description {
	return Description{
		wasSet: true,
		value:  v,
	}
}

func (b Description) String() string {
	return b.value
}

func (b Description) StringWithoutNewlines() string {
	return strings.ReplaceAll(b.value, "\n", " ")
}

func (b *Description) TodoSetManyCatgutStrings(
	vs ...*catgut.String,
) (err error) {
	var s catgut.String

	if _, err = s.Append(vs...); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return b.Set(s.String())
}

func (b *Description) TodoSetSlice(v catgut.Slice) (err error) {
	return b.Set(v.String())
}

func (b *Description) readFromRuneScannerAfterNewline(
	rs *doddish.Scanner,
	sb *strings.Builder,
) (err error) {
	if !rs.ConsumeSpacesOrErrorOnFalse() {
		return err
	}

	var r rune

	r, _, err = rs.ReadRune()
	isEOF := err == io.EOF

	if err != nil && !isEOF {
		err = errors.Wrap(err)
		return err
	}

	if r == '-' || r == '%' || r == '#' {
		if err = rs.UnreadRune(); err != nil {
			err = errors.Wrap(err)
			return err
		}

		return err
	}

	sb.WriteRune(' ')
	sb.WriteRune(r)

	if !rs.ConsumeSpacesOrErrorOnFalse() {
		return err
	}

	if err = b.readFromRuneScannerOrdinary(rs, sb); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (b *Description) readFromRuneScannerOrdinary(
	rs *doddish.Scanner,
	sb *strings.Builder,
) (err error) {
	for {
		var r rune

		r, _, err = rs.ReadRune()
		isEOF := err == io.EOF

		if err != nil && !isEOF {
			err = errors.Wrap(err)
			return err
		}

		if r == '\n' {
			if err = b.readFromRuneScannerAfterNewline(rs, sb); err != nil {
				err = errors.Wrap(err)
				return err
			}

			break
		}

		if isEOF {
			err = nil
			if r != utf8.RuneError {
				sb.WriteRune(r)
			}

			break
		}

		sb.WriteRune(r)
	}

	return err
}

func (b *Description) ReadFromBoxScanner(rs *doddish.Scanner) (err error) {
	var sb strings.Builder

	if err = b.readFromRuneScannerOrdinary(rs, &sb); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = b.Set(sb.String()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (b *Description) Set(v string) (err error) {
	b.wasSet = true

	v1 := strings.TrimSpace(v)

	if v0 := b.String(); v0 != "" && v0 != v1 {
		b.value = v0 + " " + v1
	} else {
		b.value = v1
	}

	return err
}

func (a Description) WasSet() bool {
	return a.wasSet
}

func (a *Description) Reset() {
	a.wasSet = false
	a.value = ""
}

func (a Description) IsEmpty() bool {
	return a.value == ""
}

func (a Description) Equals(b Description) (ok bool) {
	// if !a.wasSet {
	// 	return false
	// }

	return a.value == b.value
}

func (a Description) Less(b Description) (ok bool) {
	return a.value < b.value
}

func (a Description) MarshalBinary() (text []byte, err error) {
	text = []byte(a.value)
	return text, err
}

func (a *Description) UnmarshalBinary(text []byte) (err error) {
	a.wasSet = true
	a.value = string(text)
	return err
}

func (a Description) MarshalText() (text []byte, err error) {
	text = []byte(a.value)
	return text, err
}

func (a *Description) UnmarshalText(text []byte) (err error) {
	a.wasSet = true
	a.value = string(text)
	return err
}
