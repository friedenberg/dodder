package box

const (
	OpOr            = ','
	OpAnd           = ' '
	OpGroupOpen     = '['
	OpGroupClose    = ']'
	OpNegation      = '^'
	OpExact         = '='
	OpNewline       = '\n'
	OpSigilLatest   = ':'
	OpSigilHistory  = '+'
	OpSigilExternal = '.'
	OpSigilHidden   = '?'
	OpPathSeparator = '/'
	OpType          = '!'
	OpVirtual       = '%'
	OpBlob          = '@'
)

var mapOperators = map[rune]bool{
	OpOr:            true,
	OpAnd:           true,
	OpGroupOpen:     true,
	OpGroupClose:    true,
	OpNegation:      true,
	OpExact:         true,
	OpNewline:       true,
	OpSigilLatest:   true,
	OpSigilHistory:  true,
	OpSigilExternal: true,
	OpSigilHidden:   true,
}

func IsOperator(r rune, dotAllowed bool) (ok bool) {
	if dotAllowed && r == '.' {
		return
	}

	_, ok = mapOperators[r]
	return
}

var mapSequenceOperators = map[rune]bool{
	OpSigilExternal: true,
	OpPathSeparator: true,
	OpType:          true,
	OpExact:         true,
	OpVirtual:       true,
	OpBlob:          true,
}

func IsSequenceOperator(r rune) (ok bool) {
	_, ok = mapSequenceOperators[r]
	return
}
