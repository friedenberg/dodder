package doddish

import (
	"bytes"
	"io"
	"unicode"
	"unicode/utf8"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type Scanner struct {
	RuneScanner io.RuneScanner

	tokenTypeProbably TokenType

	scanned       bytes.Buffer
	scannedOffset int
	seq           Seq

	err      error
	unscan   *SeqRuneScanner
	n        int64
	lastRune rune
}

func MakeScanner(runeScanner io.RuneScanner) *Scanner {
	var scanner Scanner
	scanner.Reset(runeScanner)
	return &scanner
}

func (ts *Scanner) Reset(r io.RuneScanner) {
	ts.RuneScanner = r
	ts.scanned.Reset()
	ts.scannedOffset = 0
	ts.tokenTypeProbably = TokenTypeIncomplete
	ts.seq.Reset()
	ts.err = nil
	ts.unscan = nil
	ts.n = 0
}

func (ts *Scanner) ReadRune() (r rune, n int, err error) {
	if ts.unscan != nil {
		r, n, err = ts.unscan.ReadRune()

		if err == io.EOF {
			ts.unscan = nil
			// pass
		} else if err != nil {
			err = errors.Wrap(err)
			return
		} else {
			return
		}
	}

	ts.lastRune, n, err = ts.RuneScanner.ReadRune()
	ts.n += int64(n)

	return ts.lastRune, n, err
}

// TODO remove unread entirely
func (ts *Scanner) UnreadRune() (err error) {
	if ts.unscan != nil {
		if err = ts.unscan.UnreadRune(); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}

	err = ts.RuneScanner.UnreadRune()

	if err == nil {
		ts.n -= int64(utf8.RuneLen(ts.lastRune))
	}

	return
}

func (ts *Scanner) Unscan() {
	ts.unscan = &SeqRuneScanner{Seq: ts.seq}
}

func (ts *Scanner) CanScan() (ok bool) {
	return ts.unscan != nil || ts.err == nil
}

func (scanner *Scanner) resetBeforeNextScan() {
	scanner.scanned.Reset()
	scanner.scannedOffset = 0
	scanner.tokenTypeProbably = TokenTypeIncomplete
	scanner.seq.Reset()
}

// func (scanner *Scanner) ScanIdentifierLikeSkipSpaces() (ok bool) {
// 	if len(scanner.unscan) > 0 {
// 		ok = true
// 		scanner.unscan = nil
// 		return
// 	}

// 	if scanner.err == io.EOF {
// 		return
// 	}

// 	afterFirst := false
// 	ok = true

// 	scanner.resetBeforeNextScan()

// 	for {
// 		var r rune

// 		r, _, scanner.err = scanner.ReadRune()

// 		if scanner.err != nil {
// 			if scanner.err == io.EOF {
// 				ok = scanner.scanned.Len() > 0
// 				scanner.seq.Add(scanner.tokenTypeProbably, scanner.scanned.Bytes())
// 			}

// 			return
// 		}

// 		isOperator := unicode.IsSpace(r) || r == '[' || r == ']'
// 		isSpace := unicode.IsSpace(r)

// 		switch {
// 		case r == '"' || r == '\'':
// 			if !scanner.consumeLiteralOrFieldValue(
// 				r,
// 				TokenTypeLiteral,
// 			) {
// 				ok = false
// 				return
// 			}

// 			return

// 		case !afterFirst && isOperator:
// 			if isSpace {
// 				if !scanner.ConsumeSpacesOrErrorOnFalse() {
// 					ok = false
// 					return
// 				}

// 				continue
// 			} else {
// 				scanner.scanned.WriteRune(r)
// 				scanner.appendTokenWithTypeToSeq(TokenTypeOperator)
// 				return
// 			}

// 		case !isOperator:
// 			scanner.tokenTypeProbably = TokenTypeIdentifier
// 			scanner.scanned.WriteRune(r)
// 			afterFirst = true
// 			continue

// 		default: // wasSplitRune && afterFirst
// 			scanner.appendTokenWithTypeToSeq(scanner.tokenTypeProbably)

// 			if r == '=' {
// 				if !scanner.consumeField(r) {
// 					ok = false
// 					return
// 				}

// 				return
// 			}

// 			if scanner.err = scanner.UnreadRune(); scanner.err != nil {
// 				scanner.err = errors.Wrapf(scanner.err, "%c", r)
// 				ok = false
// 			}

// 			return
// 		}
// 	}
// }

// func (ts *Scanner) ScanSkipSpace() (ok bool) {
// 	if !ts.ConsumeSpacesOrErrorOnFalse() {
// 		return
// 	}

// 	ok = ts.Scan()

// 	return
// }

func (ts *Scanner) ScanSkipSpace() (ok bool) {
	if !ts.ConsumeSpacesOrErrorOnFalse() {
		return
	}

	ok = ts.Scan()

	return
}

func (ts *Scanner) Scan() (ok bool) {
	return ts.scan(true)
}

func (ts *Scanner) ScanDotAllowedInIdentifiers() (ok bool) {
	return ts.scan(false)
}

func (scanner *Scanner) appendTokenWithTypeToSeq(tokenType TokenType) {
	if b := scanner.scanned.Bytes()[scanner.scannedOffset:]; len(b) > 0 {
		scanner.seq.Add(tokenType, b)
		scanner.scannedOffset += len(b)
	}
}

func (scanner *Scanner) scan(dotOperatorAsSplit bool) (ok bool) {
	if scanner.unscan.IsFull() {
		ok = true
		scanner.unscan = nil
		return
	}

	if scanner.err == io.EOF {
		return
	}

	afterFirst := false
	ok = true

	scanner.resetBeforeNextScan()

	for {
		var r rune

		r, _, scanner.err = scanner.ReadRune()

		if scanner.err != nil {
			if scanner.err == io.EOF {
				ok = scanner.scanned.Len() > 0
				scanner.appendTokenWithTypeToSeq(scanner.tokenTypeProbably)
			}

			return
		}

		isOperator := IsOperator(r, !dotOperatorAsSplit)
		isSequenceOperator := IsSequenceOperator(r)
		isSpace := unicode.IsSpace(r)

		switch {
		case r == '"' || r == '\'':
			if !scanner.consumeLiteralOrFieldValue(
				r,
				TokenTypeLiteral,
			) {
				ok = false
				return
			}

			return

		case !afterFirst && isOperator:
			scanner.scanned.WriteRune(r)
			scanner.appendTokenWithTypeToSeq(TokenTypeOperator)

			if isSpace {
				if !scanner.ConsumeSpacesOrErrorOnFalse() {
					ok = false
					return
				}
			}

			return

		case !isOperator && !isSequenceOperator:
			scanner.tokenTypeProbably = TokenTypeIdentifier
			scanner.scanned.WriteRune(r)
			afterFirst = true
			continue

		case IsSequenceOperator(r):
			scanner.appendTokenWithTypeToSeq(scanner.tokenTypeProbably)
			scanner.scanned.WriteRune(r)
			scanner.appendTokenWithTypeToSeq(TokenTypeOperator)
			afterFirst = true
			continue

		default: // isOperator && afterFirst
			scanner.appendTokenWithTypeToSeq(scanner.tokenTypeProbably)

			if r == '=' {
				if !scanner.consumeField(r) {
					ok = false
					return
				}

				return
			}

			if scanner.err = scanner.UnreadRune(); scanner.err != nil {
				scanner.err = errors.Wrapf(scanner.err, "%c", r)
				ok = false
			}

			return
		}
	}
}

// Consumes any spaces currently available in the underlying RuneReader. If this
// returns false, it means that a read error has occurred, not that no spaces
// were consumed.
func (ts *Scanner) ConsumeSpacesOrErrorOnFalse() (ok bool) {
	ok = true

	for {
		var r rune

		r, _, ts.err = ts.ReadRune()

		if ts.err != nil {
			ok = false
			return
		}

		if unicode.IsSpace(r) {
			continue
		}

		if ts.err = ts.UnreadRune(); ts.err != nil {
			ok = false
			ts.err = errors.Wrapf(ts.err, "%c", r)
		}

		return
	}
}

// TODO add support for ellipis
func (scanner *Scanner) consumeLiteralOrFieldValue(
	start rune,
	tt TokenType,
) (ok bool) {
	ok = true

	lastWasBackslash := false

	for {
		var r rune

		r, _, scanner.err = scanner.ReadRune()

		if scanner.err != nil {
			ok = false
			return
		}

		currentIsBackslash := r == '\\'
		escaped := lastWasBackslash && !currentIsBackslash
		end := r == start
		content := !lastWasBackslash && !currentIsBackslash && !end

		if escaped || content {
			scanner.scanned.WriteRune(r)
		}

		if r != start || lastWasBackslash {
			lastWasBackslash = currentIsBackslash
			continue
		}

		scanner.appendTokenWithTypeToSeq(tt)

		return
	}
}

func (ts *Scanner) consumeField(start rune) bool {
	ts.scanned.WriteRune(start)
	ok := ts.consumeIdentifierLike(TokenTypeLiteral)
	return ok
}

// TODO add support for ellipsis
func (ts *Scanner) consumeIdentifierLike(
	tt TokenType,
) (ok bool) {
	ok = true

	idx := ts.scanned.Len()

	for {
		var r rune

		r, _, ts.err = ts.ReadRune()

		if ts.err != nil {
			if ts.err == io.EOF {
				ok = ts.scanned.Len() > 0
			}

			return
		}

		isOperator := IsOperator(r, true)

		switch {
		case r == '"' || r == '\'':
			if !ts.consumeLiteralOrFieldValue(r, tt) {
				ok = false
				return
			}

			return

		case !isOperator:
			ts.scanned.WriteRune(r)
			continue

		default: // wasSplitRune && afterFirst
			ts.seq.Add(
				tt,
				ts.scanned.Bytes()[idx:ts.scanned.Len()],
			)

			if ts.err = ts.UnreadRune(); ts.err != nil {
				ts.err = errors.Wrapf(ts.err, "%c", r)
				ok = false
			}

			return
		}
	}
}

// Valid only until the next call to any scan method. To keep the sequence, make
// a clone of it by calling Seq.Clone()
func (scanner *Scanner) GetSeq() Seq {
	return scanner.seq
}

func (scanner *Scanner) N() int64 {
	return scanner.n
}

func (scanner *Scanner) Error() error {
	if scanner.err == io.EOF {
		return nil
	}

	return scanner.err
}
